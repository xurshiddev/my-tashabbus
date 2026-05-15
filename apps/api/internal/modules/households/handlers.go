package households

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/http/response"
	"github.com/my-tashabbus/api/internal/modules/streets"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Handler struct {
	service *Service
	log     *slog.Logger
}

func NewHandler(service *Service, loggers ...*slog.Logger) Handler {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	if len(loggers) > 0 && loggers[0] != nil {
		log = loggers[0]
	}
	return Handler{service: service, log: log}
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	current, streetID, ok := currentUserAndID(w, r, "streetID")
	if !ok {
		return
	}
	var input CreateHouseholdInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	item, err := h.service.Create(r.Context(), current, streetID, input)
	if err != nil {
		h.writeError(w, "create_household", err)
		return
	}
	response.Data(w, http.StatusCreated, item)
}

func (h Handler) ListByStreet(w http.ResponseWriter, r *http.Request) {
	current, streetID, ok := currentUserAndID(w, r, "streetID")
	if !ok {
		return
	}
	limit := clamp(parseIntDefault(r.URL.Query().Get("limit"), 50), 1, 200)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	items, err := h.service.ListByStreet(r.Context(), current, streetID, limit, offset)
	if err != nil {
		h.writeError(w, "list_households_by_street", err)
		return
	}
	response.Paginated(w, http.StatusOK, items, limit, offset)
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.service.Get(r.Context(), current, id)
	if err != nil {
		h.writeError(w, "get_household", err)
		return
	}
	response.Data(w, http.StatusOK, item)
}

func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	var input UpdateHouseholdInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	item, err := h.service.Update(r.Context(), current, id, input)
	if err != nil {
		h.writeError(w, "update_household", err)
		return
	}
	response.Data(w, http.StatusOK, item)
}

func (h Handler) MyHouseholds(w http.ResponseWriter, r *http.Request) {
	current, ok := currentUser(w, r)
	if !ok {
		return
	}
	limit := clamp(parseIntDefault(r.URL.Query().Get("limit"), 50), 1, 200)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	items, err := h.service.MyHouseholds(r.Context(), current, limit, offset)
	if err != nil {
		h.writeError(w, "my_households", err)
		return
	}
	response.Paginated(w, http.StatusOK, items, limit, offset)
}

func (h Handler) Logs(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	limit := clamp(parseIntDefault(r.URL.Query().Get("limit"), 50), 1, 200)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	items, err := h.service.Logs(r.Context(), current, id, limit, offset)
	if err != nil {
		h.writeError(w, "household_logs", err)
		return
	}
	response.Paginated(w, http.StatusOK, items, limit, offset)
}

func currentUser(w http.ResponseWriter, r *http.Request) (users.User, bool) {
	user, err := requestcontext.CurrentUser(r.Context())
	if err != nil {
		response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing current user")
		return users.User{}, false
	}
	return user, true
}

func currentUserAndID(w http.ResponseWriter, r *http.Request, param string) (users.User, uuid.UUID, bool) {
	user, ok := currentUser(w, r)
	if !ok {
		return users.User{}, uuid.UUID{}, false
	}
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid id")
		return users.User{}, uuid.UUID{}, false
	}
	return user, id, true
}

func parseIntDefault(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (h Handler) writeError(w http.ResponseWriter, operation string, err error) {
	switch {
	case errors.Is(err, ErrForbidden), errors.Is(err, streets.ErrForbidden):
		response.ErrorCode(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource")
	case errors.Is(err, ErrHouseholdNotFound), errors.Is(err, streets.ErrStreetNotFound):
		response.ErrorCode(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, ErrDuplicateHousehold):
		response.ErrorCode(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, ErrHouseNumberRequired), errors.Is(err, ErrInvalidCounts), errors.Is(err, ErrInvalidStatus), errors.Is(err, ErrInvalidPagination):
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		h.log.Error("household request failed", "operation", operation, "error", err)
		response.ErrorCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
