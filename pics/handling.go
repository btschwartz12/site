package pics

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/btschwartz12/site/internal/repo"
	"github.com/btschwartz12/site/internal/slack"
	"github.com/btschwartz12/site/pics/assets"
	"github.com/go-chi/chi/v5"
)

var (
	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"formatRFC3339": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
	}).ParseFS(
		assets.Templates,
		"templates/base.html.tmpl",
	))
)

type templateData struct {
	Pictures []repo.Picture
	Order    string
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	pictures, err := s.rpo.GetAllPictures(r.Context())
	if err != nil {
		s.logger.Errorw("error getting pictures", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	order := "dsc"
	if r.Method == http.MethodPost {
		v := r.FormValue("order")
		if v == "asc" {
			order = v
		}
	}

	if order == "dsc" {
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Pit.After(pictures[j].Pit)
		})
	} else if order == "asc" {
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Pit.Before(pictures[j].Pit)
		})
	}

	for i, p := range pictures {
		pictures[i].Url = "/pics/static/pic/" + strconv.FormatInt(p.ID, 10) + p.Extension
	}

	templateData := templateData{
		Pictures: pictures,
		Order:    order,
	}

	if err := tmpl.ExecuteTemplate(w, "base.html.tmpl", templateData); err != nil {
		s.logger.Errorw("error executing template", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) servePictureHandler(w http.ResponseWriter, r *http.Request) {
	basename := chi.URLParam(r, "basename")

	p, err := s.rpo.GetPicture(r.Context(), basename)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "invalid extension") {
			http.Error(w, "Invalid Extension", http.StatusBadRequest)
			return
		}
		s.logger.Errorw("error getting picture", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err := os.Stat(p.Url); err != nil {
		s.logger.Errorw("error getting picture", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, p.Url)
}

func (s *server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "description is required", http.StatusBadRequest)
		return
	}

	author := r.FormValue("author")
	if author == "" {
		http.Error(w, "author is required", http.StatusBadRequest)
		return
	}

	_, err = s.rpo.InsertPicture(r.Context(), file, header, author, description)
	if err != nil {
		if strings.Contains(err.Error(), "invalid extension") {
			http.Error(w, "Invalid Extension", http.StatusBadRequest)
			return
		}
		if strings.Contains(err.Error(), "file too large") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.logger.Errorw("error inserting picture", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go s.rpo.RecordVisitor(context.Background(), r, "uploaded picture", []slack.Block{})

	http.Redirect(w, r, "/pics", http.StatusSeeOther)
}
