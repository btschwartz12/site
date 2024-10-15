package survey

import (
	"fmt"

	env "github.com/Netflix/go-env"
)

type config struct {
	Tls bool `env:"TLS" envDefault:"false"`
}

func newConfig() (*config, error) {
	conf := config{}
	if _, err := env.UnmarshalFromEnviron(&conf); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &conf, nil
}
