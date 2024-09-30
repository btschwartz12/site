package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/api/swagger"
	"github.com/btschwartz12/site/internal/repo"
)

type server struct {
	logger *zap.SugaredLogger
	rpo    *repo.Repo
	token  string
}

func NewServer(logger *zap.SugaredLogger, rpo *repo.Repo, config *Config) (*server, chi.Router, error) {
	s := &server{
		logger: logger,
		rpo:    rpo,
		token:  config.Token,
	}

	r := chi.NewRouter()
	r.Get("/", http.RedirectHandler("/api/swagger/index.html", http.StatusFound).ServeHTTP)
	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(swagger.SwaggerJSON)
	})
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/api/swagger.json")))

	r.Group(func(r chi.Router) {
		r.Use(s.tokenMiddleware)
		r.Get("/visitors", s.getVisitorsHandler)
		r.Get("/pics", s.getPicturesHandler)
		r.Post("/pics/upload", s.uploadPictureHandler)
		r.Delete("/pics/delete/{id}", s.deletePictureHandler)
		r.Put("/pics/update_likes/{id}", s.updateLikesHandler)
	})

	return s, r, nil
}
