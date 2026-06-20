package dto

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Status string `json:"status"`
}
