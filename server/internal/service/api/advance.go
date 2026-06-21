package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
	"kakeibo/internal/service/api/dto"
)

func (s *Server) handleAdvances(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleAdvances: %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		s.getAdvances(w, r)
	case http.MethodPost:
		s.createAdvance(w, r)
	case http.MethodPut:
		s.updateAdvance(w, r)
	default:
		writeError(w, 405, "method not allowed")
	}
}

// 例: /advances?transaction_id=12
func (s *Server) getAdvances(w http.ResponseWriter, r *http.Request) {
	transactionIDStr := r.URL.Query().Get("transaction_id")
	if transactionIDStr == "" {
		writeError(w, 400, "transaction_id is required")
		return
	}

	transactionID, err := strconv.ParseInt(transactionIDStr, 10, 64)
	if err != nil {
		writeError(w, 400, "invalid transaction_id")
		return
	}

	list, err := repository.GetAdvancesByTransactionID(s.DB, transactionID)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	var res []dto.AdvanceResponse
	for _, a := range list {
		res = append(res, dto.AdvanceResponse{
			ID:             a.AdvanceID,
			Name:           a.Name,
			Amount:         a.Amount,
			ReturnedAmount: a.ReturnedAmount,
			Status:         a.Status,
		})
	}

	writeJSON(w, 200, res)
}

func (s *Server) createAdvance(w http.ResponseWriter, r *http.Request) {
	writeError(w, 501, "not implemented")
}

func (s *Server) updateAdvance(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateAdvanceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	log.Printf("📦 update advance req: %+v", req)

	if req.ID <= 0 {
		writeError(w, 400, "id is required")
		return
	}
	if req.Name == "" {
		writeError(w, 400, "name is required")
		return
	}
	if req.Amount <= 0 {
		writeError(w, 400, "amount must be positive")
		return
	}

	a := model.Advance{
		AdvanceID: req.ID,
		Name:      req.Name,
		Amount:    req.Amount,
	}

	if err := repository.UpdateAdvance(s.DB, &a); err != nil {
		log.Printf("❌ UPDATE ADVANCE ERROR: %v", err)
		writeError(w, 500, err.Error())
		return
	}

	writeSuccess(w, 200, "updated")
}
