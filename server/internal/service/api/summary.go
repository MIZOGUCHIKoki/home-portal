package api

import (
	"net/http"
	"strconv"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
	"kakeibo/internal/service/api/dto"
)

func toMonthlySummaryResponse(s model.MonthlySummary) dto.MonthlySummaryResponse {
	return dto.MonthlySummaryResponse{
		YearMonth: s.YearMonth,
		Income:    s.Income,
		Expense:   s.Expense,
		Net:       s.Income - s.Expense,
	}
}

func toCategorySummaryResponse(s model.CategorySummary) dto.CategorySummaryResponse {
	return dto.CategorySummaryResponse{
		CategoryID:   s.CategoryID,
		Identifier:   s.Identifier,
		CategoryName: s.CategoryName,
		Income:       s.Income,
		Expense:      s.Expense,
		Count:        s.Count,
	}
}

// parseIntQuery parses a query parameter as *int. Returns nil (no filter)
// when the parameter is absent, and an error when it's present but invalid.
func parseIntQuery(r *http.Request, key string) (*int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (s *Server) handleMonthlySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "method not allowed")
		return
	}

	year, err := parseIntQuery(r, "year")
	if err != nil {
		writeError(w, 400, "invalid year")
		return
	}

	userID := int64(1)

	list, err := repository.GetMonthlyTransactionSummary(s.DB, userID, year)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	res := make([]dto.MonthlySummaryResponse, len(list))
	for i, item := range list {
		res[i] = toMonthlySummaryResponse(item)
	}

	writeJSON(w, 200, res)
}

func (s *Server) handleCategorySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "method not allowed")
		return
	}

	year, err := parseIntQuery(r, "year")
	if err != nil {
		writeError(w, 400, "invalid year")
		return
	}

	month, err := parseIntQuery(r, "month")
	if err != nil {
		writeError(w, 400, "invalid month")
		return
	}
	if month != nil && (*month < 1 || *month > 12) {
		writeError(w, 400, "month must be between 1 and 12")
		return
	}
	if month != nil && year == nil {
		writeError(w, 400, "year is required when month is given")
		return
	}

	userID := int64(1)

	list, err := repository.GetCategoryTransactionSummary(s.DB, userID, year, month)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	res := make([]dto.CategorySummaryResponse, len(list))
	for i, item := range list {
		res[i] = toCategorySummaryResponse(item)
	}

	writeJSON(w, 200, res)
}
