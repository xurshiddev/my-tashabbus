package handlers

import (
	"net/http"

	"github.com/my-tashabbus/api/internal/http/response"
)

type HealthHandler struct {
	ServiceName string
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: h.ServiceName,
	})
}
