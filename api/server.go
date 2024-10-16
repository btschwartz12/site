package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/api/swagger"
	"github.com/btschwartz12/site/internal/repo"
)

type ApiServer struct {
	router     *chi.Mux
	mountPoint string
}

type handler struct {
	logger *zap.SugaredLogger
	rpo    *repo.Repo
	token  string
}

func (s *ApiServer) Init(mountPoint string, logger *zap.SugaredLogger, rpo *repo.Repo) error {
	s.mountPoint = mountPoint
	s.router = chi.NewRouter()

	config, err := newConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	h := handler{
		logger: logger,
		rpo:    rpo,
		token:  config.Token,
	}

	s.router.Get("/", http.RedirectHandler("/api/swagger/index.html", http.StatusFound).ServeHTTP)
	s.router.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(swagger.SwaggerJSON)
	})
	s.router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/api/swagger.json")))

	s.router.Group(func(r chi.Router) {
		r.Use(h.tokenMiddleware)
		r.Get("/visitors", h.getVisitorsHandler)
		r.Get("/pics", h.getPicturesHandler)
		r.Post("/pics/upload", h.uploadPictureHandler)
		r.Delete("/pics/delete/{id}", h.deletePictureHandler)
		r.Put("/pics/update_likes/{id}", h.updateLikesHandler)
		r.Post("/drive/upload", h.uploadFileHandler)
		r.Get("/drive/files", h.getFilesHandler)
		r.Get("/drive/files/{id}", h.getFileHandler)
		r.Put("/drive/files/{id}", h.updateFileExpiresHandler)
	})

	return nil
}

func (s *ApiServer) GetRouter() chi.Router {
	return s.router
}

func (s *ApiServer) GetMountPoint() string {
	return s.mountPoint
}
