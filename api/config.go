package api

import (
	"fmt"

	env "github.com/Netflix/go-env"
)

type config struct {
	Token string `env:"API_TOKEN"`
}

func newConfig() (*config, error) {
	conf := config{}
	if _, err := env.UnmarshalFromEnviron(&conf); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &conf, nil
}
