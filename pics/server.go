package pics

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/internal/repo"
)

type PicsServer struct {
	logger     *zap.SugaredLogger
	rpo        *repo.Repo
	router     *chi.Mux
	mountPoint string
}

func (s *PicsServer) Init(mountPoint string, logger *zap.SugaredLogger, rpo *repo.Repo) error {
	s.logger = logger
	s.rpo = rpo
	s.mountPoint = mountPoint
	s.router = chi.NewRouter()

	s.router.HandleFunc("/", s.indexHandler)
	s.router.Post("/upload", s.uploadHandler)
	s.router.Post("/like/{id}", s.likeHandler)
	s.router.Post("/dislike/{id}", s.dislikeHandler)
	s.router.HandleFunc("/static/pic/{basename}", s.servePictureHandler)

	return nil
}

func (s *PicsServer) GetRouter() chi.Router {
	return s.router
}

func (s *PicsServer) GetMountPoint() string {
	return s.mountPoint
}
