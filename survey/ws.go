package survey

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net/http"
	"time"

	"github.com/btschwartz12/site/internal/slack"
	"github.com/gorilla/websocket"
)

type messageType byte

const (
	surveyUpdateCode messageType = iota
	numConnectionsCode
)

// handleWsMessages will run in its own goroutine, waiting for messages
// to be added to the messageQueue, then broadcasting them to all
// connected clients.
func (s *server) handleWsMessages() {
	// handle survey updates
	go func() {
		for {
			update := <-s.surveyMessageQueue
			surveyUpdateMessage := getSurveyUpdateMessage(update)
			for client := range s.clients {
				if err := client.WriteMessage(websocket.BinaryMessage, surveyUpdateMessage); err != nil {
					s.logger.Errorw("error broadcasting message", "error", err)
					client.Close()
					delete(s.clients, client)
				}
			}
		}
	}()
	// handle number of connections updates
	go func() {
		for {
			numConnections := <-s.numConnectionsMessageQueue
			numConnectionsMessage := getNumConnectionsMessage(numConnections)
			for client := range s.clients {
				if err := client.WriteMessage(websocket.BinaryMessage, numConnectionsMessage); err != nil {
					s.logger.Errorw("error broadcasting message", "error", err)
					client.Close()
					delete(s.clients, client)
				}
			}
		}
	}()
}

// getSurveyUpdateMessage will return a message to send to clients
// with the current state of the survey
func getSurveyUpdateMessage(update stateUpdate) []byte {
	var buffer bytes.Buffer
	buffer.WriteByte(byte(surveyUpdateCode))
	buffer.Write(update.MarshaledSurvey)
	return buffer.Bytes()
}

// getNumConnectionsMessage will return a message to send to clients
// with the current number of connections encoded as a uint32
func getNumConnectionsMessage(update uint32) []byte {
	var buffer bytes.Buffer
	buffer.WriteByte(byte(numConnectionsCode))
	binary.Write(&buffer, binary.BigEndian, update)
	return buffer.Bytes()
}

// wsHandler is the handler for the websocket connection. It will
// upgrade the connection, send the current state to the client,
// then keep the connection open to receive updates.
func (s *server) wsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Errorw("error upgrading connection", "error", err)
		http.Error(w, "Failed to initialize websocket connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	s.clients[conn] = true

	s.stateMutex.Lock()
	data, err := s.state.marshal()
	if err != nil {
		s.logger.Errorw("error marshaling state", "error", err)
		s.stateMutex.Unlock()
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.stateMutex.Unlock()

	// send survey update
	surveyUpdateMessage := getSurveyUpdateMessage(stateUpdate{MarshaledSurvey: data})
	if err := conn.WriteMessage(websocket.BinaryMessage, surveyUpdateMessage); err != nil {
		s.logger.Errorw("error sending state to client", "error", err)
		return
	}

	// send number of connections
	numClients := uint32(len(s.clients))
	numConnectionsMessage := getNumConnectionsMessage(numClients)
	if err := conn.WriteMessage(websocket.BinaryMessage, numConnectionsMessage); err != nil {
		s.logger.Errorw("error sending number of connections to client", "error", err)
		return
	}

	// broadcast number of connections
	s.numConnectionsMessageQueue <- numClients

	go s.rpo.RecordVisitor(context.Background(), r, "survey websocket connection", []slack.Block{})

	// wait until connection is closed
	for {
		time.Sleep(1 * time.Second)
		_, _, err := conn.NextReader()
		if err != nil {
			delete(s.clients, conn)
			break
		}
	}
}

// updateHandler is the handler for updating the survey state,
// and is called any time a client makes a change to the survey.
func (s *server) updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	s.stateMutex.Lock()
	defer s.stateMutex.Unlock()

	var svy survey
	err = svy.unmarshal(data)
	if err != nil {
		s.logger.Errorw("error unmarshaling survey", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// update the state
	s.updateState(&svy)
	// broadcast the change
	s.surveyMessageQueue <- stateUpdate{MarshaledSurvey: data}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Survey received successfully"))
}

// updateState will update the server's state with the provided survey,
// This will be called with a lock held on the stateMutex.
func (s *server) updateState(newSurvey *survey) {
	for id, question := range newSurvey.questions {
		if q, ok := s.state.questions[id]; ok {
			switch q := q.(type) {
			case *multipleChoiceQuestion:
				for i, opt := range question.(*multipleChoiceQuestion).Options {
					q.Options[i].Selected = opt.Selected
				}
			case *selectAllThatApplyQuestion:
				for i, opt := range question.(*selectAllThatApplyQuestion).Options {
					q.Options[i].Selected = opt.Selected
				}
			case *textEntryQuestion:
				q.Text = question.(*textEntryQuestion).Text
			}
		}
	}
}
