package dto

type CsvUploadResponse struct {
	ObjectKey string `json:"object_key"`
	Bucket    string `json:"bucket"`
	Size      int64  `json:"size"`
}

type CsvCheckRequest struct {
	UserID           int64  `json:"user_id"`
	ObjectKey        string `json:"object_key"`
	MethodIdentifier string `json:"method_identifier"`
}

type CsvRowResponse struct {
	Date   string `json:"date"`
	Place  string `json:"place"`
	Amount int    `json:"amount"`
	Type   bool   `json:"type"`
}

type CsvMismatchResponse struct {
	Csv         CsvRowResponse      `json:"csv"`
	Transaction TransactionResponse `json:"transaction"`
}

type CsvCheckResponse struct {
	TotalRows       int                   `json:"total_rows"`
	ReconciledCount int                   `json:"reconciled_count"`
	NewCount        int                   `json:"new_count"`
	MismatchCount   int                   `json:"mismatch_count"`
	Mismatches      []CsvMismatchResponse `json:"mismatches"`
}
