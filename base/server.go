package base

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/btschwartz12/site/base/assets"
	"github.com/btschwartz12/site/internal/handling"
	"github.com/btschwartz12/site/internal/repo"
)

type BaseServer struct {
	logger     *zap.SugaredLogger
	rpo        *repo.Repo
	router     *chi.Mux
	mountPoint string
}

func (s *BaseServer) Init(mountPoint string, logger *zap.SugaredLogger, rpo *repo.Repo) error {
	s.logger = logger
	s.rpo = rpo
	s.mountPoint = mountPoint
	s.router = chi.NewRouter()

	s.router.HandleFunc("/", s.indexHandler)
	s.router.Handle("/static/*", handling.StaticHandler(http.FileServer(http.FS(assets.Static)), ""))

	// my old stuff
	s.router.Handle("/resume*", http.RedirectHandler("/static/resume.pdf", http.StatusFound))
	s.router.Handle("/portfolio", delayRedirectHandler("https://old-portfolio.btschwartz.com/portfolio"))
	s.router.Handle("/portfolio/*", delayRedirectHandler("https://old-portfolio.btschwartz.com/portfolio"))

	return nil
}

func (s *BaseServer) GetRouter() chi.Router {
	return s.router
}

func (s *BaseServer) GetMountPoint() string {
	return s.mountPoint
}
