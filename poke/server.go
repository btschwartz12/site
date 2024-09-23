package poke

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/internal/handling"
	"github.com/btschwartz12/site/internal/repo"
	"github.com/btschwartz12/site/poke/assets"
)

type server struct {
	logger *zap.SugaredLogger
	rpo    *repo.Repo
}

func NewServer(logger *zap.SugaredLogger, rpo *repo.Repo) (*server, chi.Router, error) {
	s := &server{
		logger: logger,
		rpo:    rpo,
	}
	r := chi.NewRouter()
	r.HandleFunc("/", s.indexHandler)
	r.Handle("/static/*", handling.StaticHandler(http.FileServer(http.FS(assets.Static)), "/poke"))
	return s, r, nil
}
