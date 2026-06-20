package api

import (
	"log"
	"net/http"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
	"kakeibo/internal/service/api/dto"
)

func toMethodResponse(m model.Method) dto.MethodResponse {
	return dto.MethodResponse{
		ID:   m.MethodID,
		Name: m.MethodName,
	}
}

func (s *Server) handleMethods(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		writeError(w, 405, "method not allowed")
		return
	}

	list, err := repository.ListMethods(s.DB)
	if err != nil {
		log.Printf("❌ DB ERROR (methods): %v", err)
		writeError(w, 500, err.Error())
		return
	}

	res := make([]dto.MethodResponse, len(list))
	for i, m := range list {
		res[i] = toMethodResponse(m)
	}

	writeJSON(w, 200, res)
}
