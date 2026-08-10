package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"kakeibo/internal/model"
	"kakeibo/internal/repository"
	"kakeibo/internal/service/api/dto"
	"kakeibo/internal/service/csv"
)

const maxCsvUploadSize = 32 << 20 // 32MB

func toCsvRowResponse(date string, place string, amount int, ttype bool) dto.CsvRowResponse {
	return dto.CsvRowResponse{
		Date:   date,
		Place:  place,
		Amount: amount,
		Type:   ttype,
	}
}

// handleCsvUpload accepts a multipart CSV file and stores it in MinIO/S3,
// returning the object key so it can be passed to /csv/check.
func (s *Server) handleCsvUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed")
		return
	}

	if s.S3 == nil {
		writeError(w, 500, "S3 storage is not configured")
		return
	}

	if err := r.ParseMultipartForm(maxCsvUploadSize); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "file is required")
		return
	}
	defer file.Close()

	userID := int64(1)
	if v := r.FormValue("user_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, 400, "invalid user_id")
			return
		}
		userID = id
	}

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	if len(data) == 0 {
		writeError(w, 400, "file is empty")
		return
	}

	key, err := s.S3.UploadCSV(r.Context(), userID, header.Filename, data)
	if err != nil {
		log.Printf("❌ S3 UPLOAD ERROR: %v", err)
		writeError(w, 500, err.Error())
		return
	}

	writeJSON(w, 201, dto.CsvUploadResponse{
		ObjectKey: key,
		Bucket:    s.S3.Bucket(),
		Size:      int64(len(data)),
	})
}

// handleCsvCheck downloads a previously-uploaded CSV from S3 and compares it
// against the user's manually-entered transactions. It performs no writes:
// exact date/place/amount matches are treated as already consistent and
// skipped, and rows that partially match a manual entry (2 of the 3 fields)
// but disagree on the rest are returned so the caller can review them.
func (s *Server) handleCsvCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, 405, "method not allowed")
		return
	}

	if s.S3 == nil {
		writeError(w, 500, "S3 storage is not configured")
		return
	}

	var req dto.CsvCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, err.Error())
		return
	}

	if req.ObjectKey == "" {
		writeError(w, 400, "object_key is required")
		return
	}
	if req.MethodIdentifier == "" {
		writeError(w, 400, "method_identifier is required")
		return
	}

	userID := req.UserID
	if userID == 0 {
		userID = 1
	}

	data, err := s.S3.DownloadCSV(r.Context(), req.ObjectKey)
	if err != nil {
		log.Printf("❌ S3 DOWNLOAD ERROR: %v", err)
		writeError(w, 500, err.Error())
		return
	}

	result, err := csv.CheckCsvAgainstManual(s.DB, data, userID, req.MethodIdentifier)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}

	res := dto.CsvCheckResponse{
		TotalRows:       result.TotalRows,
		ReconciledCount: result.ReconciledCount,
		NewCount:        result.NewCount,
		MismatchCount:   len(result.Mismatches),
		Mismatches:      make([]dto.CsvMismatchResponse, 0, len(result.Mismatches)),
	}

	for _, mm := range result.Mismatches {
		advances, err := repository.GetAdvancesByTransactionID(s.DB, mm.Transaction.TransactionID)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		if advances == nil {
			advances = []model.Advance{}
		}

		res.Mismatches = append(res.Mismatches, dto.CsvMismatchResponse{
			Csv:         toCsvRowResponse(mm.CsvDate.Format("2006-01-02"), mm.CsvPlace, mm.CsvAmount, mm.CsvType),
			Transaction: toTransactionResponse(mm.Transaction, advances),
		})
	}

	writeJSON(w, 200, res)
}
