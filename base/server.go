package base

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/base/assets"
	"github.com/btschwartz12/site/internal/handling"
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
	r.Handle("/static/*", handling.StaticHandler(http.FileServer(http.FS(assets.Static)), ""))

	// my old stuff
	r.Handle("/resume*", http.RedirectHandler("/static/resume.pdf", http.StatusFound))
	r.Handle("/portfolio", delayRedirectHandler("https://old-portfolio.btschwartz.com/portfolio"))
	r.Handle("/portfolio/*", delayRedirectHandler("https://old-portfolio.btschwartz.com/portfolio"))

	return s, r, nil
}
