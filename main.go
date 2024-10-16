package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	flags "github.com/jessevdk/go-flags"

	"github.com/btschwartz12/site/api"
	"github.com/btschwartz12/site/base"
	"github.com/btschwartz12/site/drive"
	"github.com/btschwartz12/site/internal/proxy"
	"github.com/btschwartz12/site/internal/repo"
	"github.com/btschwartz12/site/pics"
	"github.com/btschwartz12/site/poke"
	"github.com/btschwartz12/site/survey"

	"go.uber.org/zap"
)

type arguments struct {
	Port        int  `short:"p" long:"port" description:"Port to listen on" default:"8080"`
	DevLogging  bool `short:"d" long:"dev-logging" description:"Enable development logging"`
	EnableProxy bool `long:"enable-proxy" description:"Enable proxying to other services"`
}

var opts arguments

func main() {
	// parse cl args
	_, err := flags.Parse(&opts)
	if err != nil {
		panic(fmt.Errorf("failed to parse flags: %w", err))
	}

	// set up logger
	var l *zap.Logger
	if opts.DevLogging {
		l, err = zap.NewDevelopment()
		if err != nil {
			panic(fmt.Errorf("failed to create logger: %w", err))
		}
	} else {
		l, err = zap.NewProduction()
		if err != nil {
			panic(fmt.Errorf("failed to create logger: %w", err))
		}
	}
	logger := l.Sugar()

	// set up repo
	rpo, err := repo.NewRepo(logger, "var")
	if err != nil {
		panic(fmt.Errorf("failed to create repo: %w", err))
	}

	// set up apps
	r := chi.NewRouter()
	apps := map[string]app{
		"/":       &base.BaseServer{},
		"/poke":   &poke.PokeServer{},
		"/survey": &survey.SurveyServer{},
		"/pics":   &pics.PicsServer{},
		"/api":    &api.ApiServer{},
		"/drive":  &drive.DriveServer{},
	}
	for mp, a := range apps {
		err = a.Init(mp, logger, rpo)
		if err != nil {
			panic(fmt.Errorf("failed to init app: %w", err))
		}
		r.Mount(a.GetMountPoint(), a.GetRouter())
	}

	// enable proxying
	if opts.EnableProxy {
		r.HandleFunc("/rust*", proxy.Proxy(os.Getenv("RUST_TARGET"), "/rust"))
		r.HandleFunc("/c*", proxy.Proxy(os.Getenv("C_TARGET"), "/c"))
	}

	// start server
	errChan := make(chan error)
	go func() {
		logger.Infow("starting server", "port", opts.Port)
		errChan <- http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), r)
	}()
	err = <-errChan
	logger.Fatalw("server error", "error", err)
}

type app interface {
	Init(mountPoint string, logger *zap.SugaredLogger, repo *repo.Repo) error
	GetRouter() chi.Router
	GetMountPoint() string
}
