package drive

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/btschwartz12/site/drive/assets"
	"github.com/go-chi/chi/v5"
)

type templateData struct {
	Title string
}

var (
	tmpl = template.Must(template.ParseFS(
		assets.Templates,
		"templates/base.html.tmpl",
	))
)

func (s *handler) indexHandler(w http.ResponseWriter, r *http.Request) {
	templateData := templateData{
		Title: "Home",
	}

	if err := tmpl.ExecuteTemplate(w, "base.html.tmpl", templateData); err != nil {
		s.logger.Errorw("error executing template", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *handler) uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		s.logger.Errorw("error getting file from form", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	notes := r.FormValue("notes")
	if notes == "" {
		http.Error(w, "notes is required", http.StatusBadRequest)
		return
	}

	f, err := s.rpo.InsertFile(r.Context(), file, header, notes)
	if err != nil {
		if strings.Contains(err.Error(), "file too large") {
			http.Error(w, "File Too Large", http.StatusBadRequest)
			return
		}
		s.logger.Errorw("error inserting file", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		ID      string `json:"file_id"`
	}{
		Success: true,
		Message: "uploaded successfully. don't lose this id!",
		ID:      f.Uuid.String(),
	}

	rsp, err := json.MarshalIndent(resp, "", " \t")
	if err != nil {
		s.logger.Errorw("error marshalling response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rsp)
}

func (s *handler) generatePermalinkHandler(w http.ResponseWriter, r *http.Request) {
	fileId := r.FormValue("file_id")
	if fileId == "" {
		http.Error(w, "file_id is required", http.StatusBadRequest)
		return
	}

	dur := r.FormValue("duration")
	if dur == "" {
		http.Error(w, "duration is required", http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(dur)
	if err != nil {
		http.Error(w, "invalid duration: must be in correct format (300s, 2h45m, etc.)", http.StatusBadRequest)
		return
	}

	p, err := s.rpo.InsertPermalink(r.Context(), fileId, int64(duration.Seconds()))
	if err != nil {
		s.logger.Errorw("error generating permalink", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		URL     string `json:"url"`
	}{
		Success: true,
		Message: "permalink generated successfully",
		URL:     fmt.Sprintf("/drive/permalinks/%s", p.Uuid),
	}

	rsp, err := json.MarshalIndent(resp, "", " \t")
	if err != nil {
		s.logger.Errorw("error marshalling response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rsp)
}

func (s *handler) servePermalinkHandler(w http.ResponseWriter, r *http.Request) {
	permalinkId := chi.URLParam(r, "permalink_id")
	if permalinkId == "" {
		http.Error(w, "permalink_id is required", http.StatusBadRequest)
		return
	}

	p, err := s.rpo.GetPermalink(r.Context(), permalinkId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "permalink not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error getting permalink", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if p.Expires.Before(time.Now()) {
		http.Error(w, fmt.Sprintf("permalink expired at %s", p.Expires), http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(p.File.Url); err != nil {
		s.logger.Errorw("error getting file", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, p.File.Url)
}
