package api

import "net/http"

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		writeError(w, 405, "method not allowed")
		return
	}

	writeSuccess(w, 200, "ok")
}
