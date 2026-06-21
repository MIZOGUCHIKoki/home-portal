package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
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

	if req.RefundAdvanceID != nil {
		t.IsTransfer = false
		t.Type = true
		t.NetAmount = t.Amount
		t.CategoryID = nil
		t.Place = nil
	}

	if len(req.Advances) > 0 {
		t.Type = false
		t.IsTransfer = false
		if t.NetAmount == 0 {
			t.NetAmount = t.Amount
		}
	}

	if t.NetAmount == 0 {
		t.NetAmount = t.Amount
	}

	return t, nil
}

func toModelTransactionFromUpdate(req dto.UpdateTransactionRequest) (model.Transaction, error) {
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
		TransactionID: req.ID,

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
	}

	var note string
	if t.Note != nil {
		note = *t.Note
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
	switch r.Method {
	case http.MethodGet:
		s.getTransactions(w, r)
	case http.MethodPost:
		s.createTransaction(w, r)
	case http.MethodPut:
		s.updateTransaction(w, r)
	case http.MethodDelete:
		s.deleteTransaction(w, r)
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

	if err := validateCreateTransactionRequest(req); err != nil {
		writeError(w, 400, err.Error())
		return
	}

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

func (s *Server) updateTransaction(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	if err := validateUpdateTransactionRequest(req); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	t, err := toModelTransactionFromUpdate(req)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}

	if err := repository.UpdateTransaction(s.DB, &t); err != nil {
		log.Printf("❌ UPDATE ERROR: %v", err)
		writeError(w, 500, err.Error())
		return
	}

	writeSuccess(w, 200, "updated")
}

// ✅ 追加
func (s *Server) deleteTransaction(w http.ResponseWriter, r *http.Request) {
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

	if err := repository.DeleteTransaction(s.DB, id); err != nil {
		log.Printf("❌ DELETE ERROR: %v", err)
		writeError(w, 500, err.Error())
		return
	}

	writeSuccess(w, 200, "deleted")
}

func validateCreateTransactionRequest(req dto.CreateTransactionRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	if req.MethodID <= 0 {
		return errors.New("method_id is required")
	}

	if req.RefundAdvanceID != nil && len(req.Advances) > 0 {
		return errors.New("refund and advances cannot be used together")
	}

	if req.RefundAdvanceID != nil {
		if req.Amount <= 0 {
			return errors.New("refund amount must be positive")
		}
		return nil
	}

	if req.NetAmount <= 0 {
		return errors.New("net_amount must be positive")
	}

	if req.NetAmount > req.Amount {
		return errors.New("net_amount must be less than or equal to amount")
	}

	if len(req.Advances) > 0 {
		sum := 0
		for _, a := range req.Advances {
			if a.Name == "" {
				return errors.New("advance name is required")
			}
			if a.Amount <= 0 {
				return errors.New("advance amount must be positive")
			}
			sum += a.Amount
		}

		if req.Amount-req.NetAmount != sum {
			return errors.New("amount - net_amount must equal total advance amount")
		}
	}

	return nil
}

func validateUpdateTransactionRequest(req dto.UpdateTransactionRequest) error {
	if req.ID <= 0 {
		return errors.New("id is required")
	}

	if req.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	if req.NetAmount <= 0 {
		return errors.New("net_amount must be positive")
	}

	if req.NetAmount > req.Amount {
		return errors.New("net_amount must be less than or equal to amount")
	}

	if req.MethodID <= 0 {
		return errors.New("method_id is required")
	}

	return nil
}
