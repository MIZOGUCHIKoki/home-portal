package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
	"kakeibo/internal/service/api/dto"
)

func toModelTransaction(req dto.CreateTransactionRequest) (model.Transaction, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return model.Transaction{}, err
	}

	var place *string
	if req.Place != "" {
		v := req.Place
		place = &v
	}

	var note *string
	if req.Note != "" {
		v := req.Note
		note = &v
	}

	t := model.Transaction{
		UserID: req.UserID,

		Date:      date,
		Amount:    req.Amount,
		NetAmount: req.NetAmount,

		Type:       req.Type,
		IsTransfer: req.IsTransfer,

		MethodID:   req.MethodID,
		CategoryID: req.CategoryID,

		Place: place,
		Note:  note,
	}

	// refund は意味を固定
	if req.RefundAdvanceID != nil {
		t.IsTransfer = false
		t.Type = true
		t.NetAmount = t.Amount
		t.CategoryID = nil
		t.Place = nil
	}

	if t.NetAmount == 0 {
		t.NetAmount = t.Amount
	}

	return t, nil
}

func toAdvanceResponse(a model.Advance) dto.AdvanceResponse {
	return dto.AdvanceResponse{
		ID:             a.AdvanceID,
		Name:           a.Name,
		Amount:         a.Amount,
		ReturnedAmount: a.ReturnedAmount,
		Status:         a.Status,
	}
}

func toTransactionResponse(t model.Transaction, advances []model.Advance) dto.TransactionResponse {
	var place string
	if t.Place != nil {
		place = *t.Place
	} else {
		place = ""
	}

	var note string
	if t.Note != nil {
		note = *t.Note
	} else {
		note = ""
	}

	res := dto.TransactionResponse{
		ID:         t.TransactionID,
		UserID:     t.UserID,
		Date:       t.Date.Format("2006-01-02"),
		Amount:     t.Amount,
		NetAmount:  t.NetAmount,
		Type:       t.Type,
		IsTransfer: t.IsTransfer,
		MethodID:   t.MethodID,
		CategoryID: t.CategoryID,
		Place:      place,
		Note:       note,
		Advances:   []dto.AdvanceResponse{},
	}

	for _, a := range advances {
		res.Advances = append(res.Advances, toAdvanceResponse(a))
	}

	return res
}

func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleTransactions: %s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		s.getTransactions(w, r)
	case http.MethodPost:
		s.createTransaction(w, r)
	default:
		writeError(w, 405, "method not allowed")
	}
}

func (s *Server) getTransactions(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	list, err := repository.GetTransactions(s.DB, userID)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	responses := make([]dto.TransactionResponse, 0, len(list))
	for _, t := range list {
		advances, err := repository.GetAdvancesByTransactionID(s.DB, t.TransactionID)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		if advances == nil {
			advances = []model.Advance{}
		}

		responses = append(responses, toTransactionResponse(t, advances))
	}

	writeJSON(w, 200, responses)
}

func (s *Server) createTransaction(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	log.Printf("📦 req: %+v", req)

	tx, err := s.DB.Begin()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	t, err := toModelTransaction(req)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}

	err = repository.CreateTransactionTx(tx, &t)
	if err != nil {
		log.Printf("❌ DB ERROR: %v", err)
		writeError(w, 500, err.Error())
		return
	}

	log.Printf("✅ transaction created: %d", t.TransactionID)

	for _, a := range req.Advances {
		if a.Name == "" || a.Amount <= 0 {
			continue
		}
		if err := repository.CreateAdvanceTx(tx, t.TransactionID, a.Name, a.Amount); err != nil {
			writeError(w, 500, err.Error())
			return
		}
	}

	if req.RefundAdvanceID != nil {
		if err := repository.ApplyRefundTx(tx, *req.RefundAdvanceID, t.Amount); err != nil {
			writeError(w, 500, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, 500, err.Error())
		return
	}

	writeSuccess(w, 201, "created")
}
