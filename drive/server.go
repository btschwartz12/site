package drive

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/internal/repo"
)

type DriveServer struct {
	router     *chi.Mux
	mountPoint string
}

type handler struct {
	logger *zap.SugaredLogger
	rpo    *repo.Repo
}

func (s *DriveServer) Init(mountPoint string, logger *zap.SugaredLogger, rpo *repo.Repo) error {
	s.mountPoint = mountPoint
	s.router = chi.NewRouter()

	h := handler{
		logger: logger,
		rpo:    rpo,
	}

	s.router.HandleFunc("/", h.indexHandler)
	s.router.Post("/upload", h.uploadFileHandler)
	s.router.Post("/generate_permalink", h.generatePermalinkHandler)
	s.router.Get("/permalinks/{permalink_id}", h.servePermalinkHandler)

	return nil
}

func (s *DriveServer) GetRouter() chi.Router {
	return s.router
}

func (s *DriveServer) GetMountPoint() string {
	return s.mountPoint
}
