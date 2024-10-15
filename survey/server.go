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
	rateLimitPerSec = 10
)

type stateUpdate struct {
	MarshaledSurvey []byte
}

type SurveyServer struct {
	logger     *zap.SugaredLogger
	rpo        *repo.Repo
	router     *chi.Mux
	mountPoint string
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

func (s *SurveyServer) Init(mountPoint string, logger *zap.SugaredLogger, rpo *repo.Repo) error {
	s.mountPoint = mountPoint
	s.logger = logger
	s.rpo = rpo
	s.router = chi.NewRouter()

	s.clients = make(map[*websocket.Conn]bool)
	s.surveyMessageQueue = make(chan stateUpdate, maxQueueSize)
	s.numConnectionsMessageQueue = make(chan uint32, maxQueueSize)

	config, err := newConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	s.tls = config.Tls

	s.state, err = parseSurveyFromYAML(assets.SurveyYAML)
	if err != nil {
		return fmt.Errorf("failed to parse survey.yaml: %w", err)
	}

	s.router.Group(func(r chi.Router) {
		r.Use(httprate.Limit(
			rateLimitPerSec,
			1*time.Second,
			httprate.WithKeyFuncs(httprate.KeyByIP),
		))
		r.HandleFunc("/update", s.updateHandler)
	})
	s.router.HandleFunc("/", s.indexHandler)
	s.router.HandleFunc("/ws", s.wsHandler)
	s.router.Handle("/static/*", handling.StaticHandler(http.FileServer(http.FS(assets.Static)), "/survey"))

	// restore state from db
	err = s.restoreState()
	if err != nil {
		return fmt.Errorf("failed to restore state: %w", err)
	}

	// start the websocket message handler
	go s.handleWsMessages()

	return nil
}

func (s *SurveyServer) restoreState() error {
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

func (s *SurveyServer) GetRouter() chi.Router {
	return s.router
}

func (s *SurveyServer) GetMountPoint() string {
	return s.mountPoint
}
