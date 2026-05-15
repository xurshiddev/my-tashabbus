package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type DataResponse struct {
	Data any `json:"data"`
}

type PaginatedResponse struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func Error(w http.ResponseWriter, status int, message string) {
	ErrorCode(w, status, "INTERNAL_ERROR", message)
}

func ErrorCode(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}

func Data(w http.ResponseWriter, status int, payload any) {
	JSON(w, status, DataResponse{Data: payload})
}

func Paginated(w http.ResponseWriter, status int, payload any, limit, offset int) {
	JSON(w, status, PaginatedResponse{
		Data: payload,
		Pagination: Pagination{
			Limit:  limit,
			Offset: offset,
		},
	})
}
