package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// uploadFileHandler godoc
// @Summary Upload a file
// @Description Upload a file
// @Tags drive
// @Accept mpfd
// @Produce json
// @Param file formData file true "File"
// @Param notes formData string true "Notes"
// @Router /api/drive/upload [post]
// @Security Bearer
// @Success 200 {object} repo.File
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

	if err := json.NewEncoder(w).Encode(f); err != nil {
		s.logger.Errorw("error encoding picture", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getFileHandler godoc
// @Summary Get a file
// @Description Get a file
// @Tags drive
// @Param id path string true "File ID"
// @Router /api/drive/files/{id} [get]
// @Security Bearer
// @Success 200 {object} repo.File
func (s *handler) getFileHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	f, err := s.rpo.GetFile(r.Context(), id)
	if err != nil {
		s.logger.Errorw("error getting file", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(f); err != nil {
		s.logger.Errorw("error encoding picture", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getFilesHandler godoc
// @Summary Get all files
// @Description Get all files
// @Tags drive
// @Router /api/drive/files [get]
// @Security Bearer
// @Success 200 {array} repo.File
func (s *handler) getFilesHandler(w http.ResponseWriter, r *http.Request) {
	files, err := s.rpo.GetAllFiles(r.Context())
	if err != nil {
		s.logger.Errorw("error getting files", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(files, "", " \t")
	if err != nil {
		s.logger.Errorw("error marshalling files", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// generatePermalinkHandler godoc
// @Summary Generate a permalink
// @Description Generate a permalink
// @Tags drive
// @Param file_id formData string true "File ID"
// @Param duration formData string true "Duration (300s, 2h45m, etc.)"
// @Router /api/drive/files/{id}/permalink [post]
// @Security Bearer
// @Success 200 {object} repo.Permalink
func (s *handler) generatePermalinkHandler(w http.ResponseWriter, r *http.Request) {
	fileId := chi.URLParam(r, "id")
	if fileId == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	dur := r.FormValue("duration_seconds")
	if dur == "" {
		http.Error(w, "duration_seconds is required", http.StatusBadRequest)
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

	resp, err := json.MarshalIndent(p, "", " \t")
	if err != nil {
		s.logger.Errorw("error marshalling permalink", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// getPermalinksHandler godoc
// @Summary Get all permalinks
// @Description Get all permalinks
// @Tags drive
// @Router /api/drive/files/permalinks [get]
// @Security Bearer
// @Success 200 {array} repo.Permalink
func (s *handler) getPermalinksHandler(w http.ResponseWriter, r *http.Request) {
	permalinks, err := s.rpo.GetAllPermalinks(r.Context())
	if err != nil {
		s.logger.Errorw("error getting permalinks", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(permalinks, "", " \t")
	if err != nil {
		s.logger.Errorw("error marshalling permalinks", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// servePermalinkHandler godoc
// @Summary Serve a permalink
// @Description Serve a permalink
// @Tags drive
// @Param id path string true "Permalink ID"
// @Router /api/drive/files/permalinks/{id}/ [get]
// @Security Bearer
// @Success 200
func (s *handler) servePermalinkHandler(w http.ResponseWriter, r *http.Request) {
	permalinkId := chi.URLParam(r, "id")
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

	if _, err := os.Stat(p.File.Url); os.IsNotExist(err) {
		s.logger.Errorw("error getting file", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, p.File.Url)
}
