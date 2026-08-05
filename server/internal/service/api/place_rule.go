package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
	"kakeibo/internal/service/api/dto"
)

func toPlaceRuleResponse(r model.PlaceRule) dto.PlaceRuleResponse {
	return dto.PlaceRuleResponse{
		ID:         r.PlaceRuleID,
		Text:       r.Text,
		MatchType:  int(r.MatchType),
		Identifier: r.Identifier,
		IsTransfer: r.IsTransfer,
	}
}

func isValidPlaceRuleMatchType(mt int) bool {
	switch model.PlaceRuleMatchType(mt) {
	case model.PlaceRuleMatchContains, model.PlaceRuleMatchExact, model.PlaceRuleMatchPrefix, model.PlaceRuleMatchSuffix:
		return true
	default:
		return false
	}
}

func (s *Server) handlePlaceRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getPlaceRules(w, r)
	case http.MethodPost:
		s.createPlaceRule(w, r)
	case http.MethodPut:
		s.updatePlaceRule(w, r)
	case http.MethodDelete:
		s.deletePlaceRule(w, r)
	default:
		writeError(w, 405, "method not allowed")
	}
}

func (s *Server) getPlaceRules(w http.ResponseWriter, r *http.Request) {
	list, err := repository.ListPlaceRules(s.DB)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	res := make([]dto.PlaceRuleResponse, len(list))
	for i, item := range list {
		res[i] = toPlaceRuleResponse(item)
	}

	writeJSON(w, 200, res)
}

func (s *Server) createPlaceRule(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePlaceRuleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	if req.Text == "" {
		writeError(w, 400, "text is required")
		return
	}
	if !isValidPlaceRuleMatchType(req.MatchType) {
		writeError(w, 400, "match_type must be 0, 1, 2 or 3")
		return
	}

	id, err := repository.CreatePlaceRule(s.DB, req.Text, model.PlaceRuleMatchType(req.MatchType), req.Identifier, req.IsTransfer)
	if err != nil {
		log.Printf("❌ DB ERROR (place_rules create): %v", err)
		writeError(w, 500, err.Error())
		return
	}

	writeJSON(w, 201, dto.PlaceRuleResponse{
		ID:         id,
		Text:       req.Text,
		MatchType:  req.MatchType,
		Identifier: req.Identifier,
		IsTransfer: req.IsTransfer,
	})
}

func (s *Server) updatePlaceRule(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePlaceRuleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	if req.ID <= 0 {
		writeError(w, 400, "id is required")
		return
	}
	if req.Text == "" {
		writeError(w, 400, "text is required")
		return
	}
	if !isValidPlaceRuleMatchType(req.MatchType) {
		writeError(w, 400, "match_type must be 0, 1, 2 or 3")
		return
	}

	if err := repository.UpdatePlaceRule(s.DB, req.ID, req.Text, model.PlaceRuleMatchType(req.MatchType), req.Identifier, req.IsTransfer); err != nil {
		log.Printf("❌ UPDATE ERROR (place_rules): %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, 404, "not found")
			return
		}
		writeError(w, 500, err.Error())
		return
	}

	writeSuccess(w, 200, "updated")
}

func (s *Server) deletePlaceRule(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeError(w, 400, "id is required")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, 400, "invalid id")
		return
	}

	if err := repository.DeletePlaceRule(s.DB, id); err != nil {
		log.Printf("❌ DELETE ERROR (place_rules): %v", err)
		writeError(w, 500, err.Error())
		return
	}

	writeSuccess(w, 200, "deleted")
}
