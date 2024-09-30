package api

import (
	"encoding/json"
	"net/http"
)

// getVisitorsHandler godoc
// @Summary Get visitors
// @Description Get the visitors
// @Tags visitors
// @Produce json
// @Router /api/visitors [get]
// @Security Bearer
// @Success 200 {array} repo.Visitor
func (s *server) getVisitorsHandler(w http.ResponseWriter, r *http.Request) {
	visitors, err := s.rpo.GetAllVisitors(r.Context())
	if err != nil {
		s.logger.Errorw("error getting visitors", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(visitors, "", " \t")
	if err != nil {
		s.logger.Errorw("error marshalling visitors", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
