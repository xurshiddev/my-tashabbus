package responsibles

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
	var input CreateAssignmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	item, err := h.service.Create(r.Context(), current, streetID, input)
	if err != nil {
		h.writeError(w, "create_responsible_assignment", err)
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
		h.writeError(w, "list_responsible_assignments", err)
		return
	}
	response.Paginated(w, http.StatusOK, items, limit, offset)
}

func (h Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	item, err := h.service.Deactivate(r.Context(), current, id)
	if err != nil {
		h.writeError(w, "deactivate_responsible_assignment", err)
		return
	}
	response.Data(w, http.StatusOK, item)
}

func currentUserAndID(w http.ResponseWriter, r *http.Request, param string) (users.User, uuid.UUID, bool) {
	user, err := requestcontext.CurrentUser(r.Context())
	if err != nil {
		response.ErrorCode(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing current user")
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
	case errors.Is(err, ErrAssignmentNotFound), errors.Is(err, streets.ErrStreetNotFound), errors.Is(err, users.ErrUserNotFound):
		response.ErrorCode(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, ErrResponsibleRoleRequired), errors.Is(err, ErrResponsibleUserInactive), errors.Is(err, ErrResponsibleWrongMFY), errors.Is(err, ErrHouseRangeRequired), errors.Is(err, ErrInvalidHouseRange), errors.Is(err, ErrInvalidPagination):
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		h.log.Error("responsible assignment request failed", "operation", operation, "error", err)
		response.ErrorCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
