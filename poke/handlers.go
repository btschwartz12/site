package poke

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/btschwartz12/site/internal/slack"
	"github.com/btschwartz12/site/poke/assets"
)

var (
	tmpl = template.Must(template.ParseFS(
		assets.Templates,
		"templates/base.html.tmpl",
	))
)

type templateData struct {
	Encounter *encounter
	ShinyOdds string
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	templateData := templateData{
		ShinyOdds: fmt.Sprintf("1 in %d", s.getShinyOddsDenom()),
	}

	if r.Method == http.MethodPost {
		templateData.Encounter = s.getEncounter()

		if templateData.Encounter.Shiny {
			go s.rpo.RecordVisitor(context.Background(), r, "shiny encounter", s.getShinyEncounterBlocks(templateData.Encounter))
		}
	}

	if err := tmpl.ExecuteTemplate(w, "base.html.tmpl", templateData); err != nil {
		s.logger.Errorw("error executing template", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) getShinyEncounterBlocks(encounter *encounter) []slack.Block {
	blocks := []slack.Block{
		{
			Type: "context",
			Elements: []slack.Element{
				{
					Type: "mrkdwn",
					Text: "got a shiny!",
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("pokedex number: %s", encounter.PokedexNumber),
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("odds: 1 in %d", encounter.ShinyDenom),
				},
			},
		},
	}
	return blocks
}
