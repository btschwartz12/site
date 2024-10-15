package base

import (
	"html/template"
	"net/http"

	"github.com/btschwartz12/site/base/assets"
)

var (
	funcMap = template.FuncMap{
		"noescape": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	tmpl = template.Must(template.New("").Funcs(funcMap).ParseFS(
		assets.Templates,
		"templates/base.html.tmpl",
	))
)

type templateData struct {
	Title string
}

func (s *BaseServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	templateData := templateData{
		Title: "Home",
	}

	if err := tmpl.ExecuteTemplate(w, "base.html.tmpl", templateData); err != nil {
		s.logger.Errorw("error executing template", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func delayRedirectHandler(destination string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Refresh", "4; url="+destination)
		msg := "you have stumbled across a legacy service.\n\ni'll redirect you to its burial site.\n\none moment please..."
		w.Write([]byte(msg))
	})
}
