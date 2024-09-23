package survey

import (
	"fmt"

	env "github.com/Netflix/go-env"
)

type Config struct {
	Tls bool `env:"TLS" envDefault:"false"`
}

func NewConfig() (*Config, error) {
	conf := Config{}
	if _, err := env.UnmarshalFromEnviron(&conf); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &conf, nil
}
