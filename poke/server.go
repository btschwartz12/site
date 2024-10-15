package poke

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/internal/handling"
	"github.com/btschwartz12/site/internal/repo"
	"github.com/btschwartz12/site/poke/assets"
)

type PokeServer struct {
	logger     *zap.SugaredLogger
	rpo        *repo.Repo
	router     *chi.Mux
	mountPoint string
}

func (s *PokeServer) Init(mountPoint string, logger *zap.SugaredLogger, rpo *repo.Repo) error {
	s.logger = logger
	s.rpo = rpo
	s.mountPoint = mountPoint
	s.router = chi.NewRouter()

	s.router.HandleFunc("/", s.indexHandler)
	s.router.Handle("/static/*", handling.StaticHandler(http.FileServer(http.FS(assets.Static)), "/poke"))
	return nil
}

func (s *PokeServer) GetRouter() chi.Router {
	return s.router
}

func (s *PokeServer) GetMountPoint() string {
	return s.mountPoint
}
