package survey

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/btschwartz12/site/survey/assets"
)

var (
	tmpl = template.Must(template.ParseFS(
		assets.Templates,
		"templates/base.html.tmpl",
		"templates/survey.html.tmpl",
	))
)

type templateData struct {
	SurveyData         surveyTemplateData
	WsProtocol         string
	SurveyUpdateCode   byte
	NumConnectionsCode byte
}

type surveyTemplateData struct {
	Version   uint8
	Questions []questionData
}

type questionData struct {
	ID      uint8
	Type    string // "multiple_choice", "select_all", "text_entry"
	Title   string
	Options []optionData // For select questions
	Text    string       // For text entry questions
}

type optionData struct {
	Index    int
	Title    string
	Selected bool
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	templateData := templateData{
		SurveyData:         s.getSurveyTemplateData(),
		SurveyUpdateCode:   byte(surveyUpdateCode),
		NumConnectionsCode: byte(numConnectionsCode),
	}

	if s.tls {
		templateData.WsProtocol = "wss"
	} else {
		templateData.WsProtocol = "ws"
	}

	if err := tmpl.ExecuteTemplate(w, "base.html.tmpl", templateData); err != nil {
		s.logger.Errorw("error executing template", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) getSurveyTemplateData() surveyTemplateData {
	templateData := surveyTemplateData{
		Version: s.state.version,
	}

	for id, question := range s.state.questions {
		qData := questionData{
			ID:    id,
			Title: question.getTitle(),
		}

		switch q := question.(type) {
		case *multipleChoiceQuestion:
			qData.Type = "multiple_choice"
			for idx, opt := range q.Options {
				qData.Options = append(qData.Options, optionData{
					Index:    idx,
					Title:    opt.Title,
					Selected: opt.Selected,
				})
			}
		case *selectAllThatApplyQuestion:
			qData.Type = "select_all"
			for idx, opt := range q.Options {
				qData.Options = append(qData.Options, optionData{
					Index:    idx,
					Title:    opt.Title,
					Selected: opt.Selected,
				})
			}
		case *textEntryQuestion:
			qData.Type = "text_entry"
			qData.Text = q.Text
		default:
			continue
		}

		templateData.Questions = append(templateData.Questions, qData)
	}

	sort.Slice(templateData.Questions, func(i, j int) bool {
		return templateData.Questions[i].ID < templateData.Questions[j].ID
	})

	return templateData
}
