package api

import (
	"encoding/json"
	"net/http"
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
// @Param expires formData string true "Expiration"
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

	expires := r.FormValue("expires")
	if expires == "" {
		http.Error(w, "expires is required", http.StatusBadRequest)
		return
	}
	t, err := time.Parse(time.RFC3339, expires)
	if err != nil {
		s.logger.Errorw("error parsing time", "error", err)
		http.Error(w, "invalid time format", http.StatusBadRequest)
		return
	}

	f, err := s.rpo.InsertFile(r.Context(), file, header, t, notes)
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

// updateFileExpiresHandler godoc
// @Summary Update a file's expiration
// @Description Update a file's expiration
// @Tags drive
// @Param id path string true "File ID"
// @Param expires formData string true "New expiration"
// @Router /api/drive/files/{id} [put]
// @Security Bearer
// @Success 200 {object} repo.File
func (s *handler) updateFileExpiresHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	expires := r.FormValue("expires")
	if expires == "" {
		http.Error(w, "expires is required", http.StatusBadRequest)
		return
	}
	t, err := time.Parse(time.RFC3339, expires)
	if err != nil {
		http.Error(w, "invalid time format", http.StatusBadRequest)
		return
	}

	f, err := s.rpo.UpdateFileExpires(r.Context(), id, t)
	if err != nil {
		s.logger.Errorw("error updating file expiration", "error", err)
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
