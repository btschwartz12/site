package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// uploadPictureHandler godoc
// @Summary Upload a picture
// @Description Upload a picture
// @Tags pictures
// @Accept mpfd
// @Produce json
// @Param file formData file true "Picture file"
// @Param description formData string true "Description of the picture"
// @Router /api/pics/upload [post]
// @Security Bearer
// @Success 200 {object} repo.Picture
func (s *server) uploadPictureHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		s.logger.Errorw("error getting file from form", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "description is required", http.StatusBadRequest)
		return
	}

	p, err := s.rpo.InsertPicture(r.Context(), file, header, description)
	if err != nil {
		if strings.Contains(err.Error(), "invalid extension") {
			http.Error(w, "Invalid Extension", http.StatusBadRequest)
			return
		}
		if strings.Contains(err.Error(), "file too large") {
			http.Error(w, "File Too Large", http.StatusBadRequest)
			return
		}
		s.logger.Errorw("error inserting picture", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(p); err != nil {
		s.logger.Errorw("error encoding picture", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// deletePictureHandler godoc
// @Summary Delete a picture
// @Description Delete a picture
// @Tags pictures
// @Param id path string true "Picture ID"
// @Router /api/pics/delete/{id} [delete]
// @Security Bearer
// @Success 204
func (s *server) deletePictureHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := s.rpo.DeletePicture(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("error deleting picture", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
