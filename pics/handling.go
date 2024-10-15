package pics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
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

func (s *PicsServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	pictures, err := s.rpo.GetAllPictures(r.Context())
	if err != nil {
		s.logger.Errorw("error getting pictures", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	order := "dsc"
	v := r.FormValue("order")
	allowedOrders := map[string]bool{
		"asc":      true,
		"dsc":      true,
		"random":   true,
		"likes":    true,
		"dislikes": true,
	}
	if allowedOrders[v] {
		order = v
	}

	switch order {
	case "dsc":
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Pit.After(pictures[j].Pit)
		})
	case "asc":
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].Pit.Before(pictures[j].Pit)
		})
	case "likes":
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].NumLikes > pictures[j].NumLikes
		})
	case "dislikes":
		sort.Slice(pictures, func(i, j int) bool {
			return pictures[i].NumDislikes > pictures[j].NumDislikes
		})
	case "random":
		rand.Shuffle(len(pictures), func(i, j int) {
			pictures[i], pictures[j] = pictures[j], pictures[i]
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

func (s *PicsServer) likeHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	order := r.FormValue("order")
	if order == "" {
		order = "dsc"
	}

	err := s.rpo.LikePicture(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error liking picture", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/pics?order=%s", order), http.StatusSeeOther)
}

func (s *PicsServer) dislikeHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	order := r.FormValue("order")
	if order == "" {
		order = "dsc"
	}

	err := s.rpo.DislikePicture(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error disliking picture", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/pics?order=%s", order), http.StatusSeeOther)
}

func (s *PicsServer) servePictureHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *PicsServer) uploadHandler(w http.ResponseWriter, r *http.Request) {
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

	go s.rpo.RecordVisitor(context.Background(), r, "uploaded picture", getPictureBlocks(author, description))

	http.Redirect(w, r, "/pics", http.StatusSeeOther)
}

func getPictureBlocks(author, description string) []slack.Block {
	blocks := []slack.Block{
		{
			Type: "context",
			Elements: []slack.Element{
				{
					Type: "mrkdwn",
					Text: "pic uploaded!",
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("author: %s", author),
				},
				{
					Type: "mrkdwn",
					Text: fmt.Sprintf("caption: %s", description),
				},
			},
		},
	}
	return blocks
}
