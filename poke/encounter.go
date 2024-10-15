package poke

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/btschwartz12/site/poke/assets"
)

const (
	normalPathPrefix = "/poke/static/img/normal/"
	shinyPathPrefix  = "/poke/static/img/shiny/"
	extension        = ".gif"
)

var (
	normalPokedexNumbers []string
	shinyPokedexNumbers  []string
)

func init() {
	entries, err := assets.Static.ReadDir("static/img/normal")
	if err != nil {
		panic(fmt.Errorf("failed to read normal sprite directory: %w", err))
	}
	for _, entry := range entries {
		base := filepath.Base(entry.Name())
		pokedexNumber := strings.TrimSuffix(base, filepath.Ext(base))
		normalPokedexNumbers = append(normalPokedexNumbers, pokedexNumber)
	}
	entries, err = assets.Static.ReadDir("static/img/shiny")
	if err != nil {
		panic(fmt.Errorf("failed to read shiny sprite directory: %w", err))
	}
	for _, entry := range entries {
		base := filepath.Base(entry.Name())
		pokedexNumber := strings.TrimSuffix(base, filepath.Ext(base))
		shinyPokedexNumbers = append(shinyPokedexNumbers, pokedexNumber)
	}
}

type encounter struct {
	PokedexNumber string
	SpritePath    string
	Shiny         bool
	ShinyDenom    int
}

func (s *PokeServer) getEncounter() *encounter {
	shinyDenom := s.getShinyOddsDenom()
	shinyChance := 1 / float64(shinyDenom)
	shiny := rand.Float64() < shinyChance
	if shiny {
		pokedexNumber := shinyPokedexNumbers[rand.Intn(len(shinyPokedexNumbers))]
		return &encounter{
			PokedexNumber: pokedexNumber,
			SpritePath:    shinyPathPrefix + pokedexNumber + extension,
			Shiny:         true,
			ShinyDenom:    shinyDenom,
		}
	} else {
		pokedexNumber := normalPokedexNumbers[rand.Intn(len(normalPokedexNumbers))]
		return &encounter{
			PokedexNumber: pokedexNumber,
			SpritePath:    normalPathPrefix + pokedexNumber + extension,
			Shiny:         false,
			ShinyDenom:    shinyDenom,
		}
	}
}

// getShinyOddsDenom returns the denominator of the odds of
// encountering a shiny pokemon.
func (s *PokeServer) getShinyOddsDenom() int {
	dailyDenoms := map[string]int{
		"Sunday":    1 << 9,
		"Monday":    1 << 8,
		"Tuesday":   1 << 7,
		"Wednesday": 1 << 10,
		"Thursday":  1 << 5,
		"Friday":    1 << 4,
		"Saturday":  1 << 6,
	}
	weekday := time.Now().Weekday().String()
	denom, ok := dailyDenoms[weekday]
	if !ok {
		s.logger.Errorw("failed to find daily denominator", "weekday", weekday)
		return 2309480209
	}
	return denom
}
