package survey

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/btschwartz12/site/internal/handling"
	"github.com/btschwartz12/site/internal/repo"
	"github.com/btschwartz12/site/survey/assets"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	maxQueueSize    = 10000
	rateLimitPerMin = 30
)

type stateUpdate struct {
	MarshaledSurvey []byte
}

type server struct {
	logger *zap.SugaredLogger
	rpo    *repo.Repo
	// needed to determine ws protocol (ws vs. wss)
	tls bool
	// state is the current state of the survey
	state      *survey
	stateMutex sync.Mutex
	// clients is a map of all connected clients
	clients map[*websocket.Conn]bool
	// surveyMessageQueue is a channel that holds incoming state updates
	surveyMessageQueue chan stateUpdate
	// numConnectionsMessageQueue is a channel that holds incoming connection count updates
	numConnectionsMessageQueue chan uint32
}

func NewServer(logger *zap.SugaredLogger, rpo *repo.Repo, config *Config) (*server, chi.Router, error) {
	s := &server{
		logger:                     logger,
		rpo:                        rpo,
		clients:                    make(map[*websocket.Conn]bool),
		surveyMessageQueue:         make(chan stateUpdate, maxQueueSize),
		numConnectionsMessageQueue: make(chan uint32, maxQueueSize),
		tls:                        config.Tls,
	}

	var err error
	s.state, err = parseSurveyFromYAML(assets.SurveyYAML)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse survey.yaml: %w", err)
	}

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			rateLimitPerMin,
			1*time.Second,
			httprate.WithKeyFuncs(httprate.KeyByIP),
		))
		r.HandleFunc("/update", s.updateHandler)
	})
	r.HandleFunc("/", s.indexHandler)
	r.HandleFunc("/ws", s.wsHandler)
	r.Handle("/static/*", handling.StaticHandler(http.FileServer(http.FS(assets.Static)), "/survey"))

	go s.handleWsMessages()

	return s, r, nil
}

func (s *server) RestoreState() error {
	existingState, err := s.rpo.GetSurveyState(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get survey state: %w", err)
	}
	err = s.updateState(existingState)
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}
	return nil
}
