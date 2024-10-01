package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	flags "github.com/jessevdk/go-flags"

	"github.com/btschwartz12/site/api"
	"github.com/btschwartz12/site/base"
	"github.com/btschwartz12/site/internal/repo"
	"github.com/btschwartz12/site/pics"
	"github.com/btschwartz12/site/poke"
	"github.com/btschwartz12/site/survey"

	"go.uber.org/zap"
)

type arguments struct {
	Port       int  `short:"p" long:"port" description:"Port to listen on" default:"8080"`
	DevLogging bool `short:"d" long:"dev" description:"Enable development logging"`
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

	// repo
	rpo, err := repo.NewRepo(logger, "var")
	if err != nil {
		panic(fmt.Errorf("failed to create repo: %w", err))
	}

	// base
	_, baseRouter, err := base.NewServer(logger, rpo)
	if err != nil {
		panic(fmt.Errorf("failed to create base server: %w", err))
	}

	// api
	apiConfig, err := api.NewConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create api config: %w", err))
	}
	_, apiRouter, err := api.NewServer(logger, rpo, apiConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create api server: %w", err))
	}

	// survey
	surveyConfig, err := survey.NewConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create survey config: %w", err))
	}
	surveyServer, surveyRouter, err := survey.NewServer(logger, rpo, surveyConfig)
	if err != nil {
		panic(fmt.Errorf("failed to create survey server: %w", err))
	}
	err = surveyServer.RestoreState()
	if err != nil {
		logger.Errorw("failed to restore survey state", "error", err)
	}

	// poke
	_, pokeRouter, err := poke.NewServer(logger, rpo)
	if err != nil {
		panic(fmt.Errorf("failed to create poke server: %w", err))
	}

	// pics
	_, picsRouter, err := pics.NewServer(logger, rpo)
	if err != nil {
		panic(fmt.Errorf("failed to create pics server: %w", err))
	}

	// do the thing
	r := chi.NewRouter()
	r.Mount("/", baseRouter)
	r.Mount("/api", apiRouter)
	r.Mount("/survey", surveyRouter)
	r.Mount("/poke", pokeRouter)
	r.Mount("/pics", picsRouter)

	errChan := make(chan error)
	go func() {
		logger.Infow("starting server", "port", opts.Port)
		errChan <- http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), r)
	}()
	err = <-errChan
	logger.Fatalw("server error", "error", err)
}
