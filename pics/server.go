package pics

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/internal/repo"
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
	r.Post("/upload", s.uploadHandler)
	r.Post("/like/{id}", s.likeHandler)
	r.Post("/dislike/{id}", s.dislikeHandler)
	r.HandleFunc("/static/pic/{basename}", s.servePictureHandler)
	return s, r, nil
}
