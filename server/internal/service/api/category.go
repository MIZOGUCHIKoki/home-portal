package api

import (
	"net/http"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
	"kakeibo/internal/service/api/dto"
)

func (s *Server) handleCategories(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		writeError(w, 405, "method not allowed")
		return
	}

	list, err := repository.ListCategories(s.DB)
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}

	categoryResponses := make([]dto.CategoryResponse, len(list))
	for i, category := range list {
		categoryResponses[i] = toCategoryResponse(category)
	}
	writeJSON(w, 200, categoryResponses)
}

func toCategoryResponse(c model.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:   c.CategoryID,
		Name: c.CategoryName,
	}
}
