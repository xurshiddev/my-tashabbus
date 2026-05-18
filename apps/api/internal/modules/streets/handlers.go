package streets

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/my-tashabbus/api/internal/http/requestcontext"
	"github.com/my-tashabbus/api/internal/http/response"
	"github.com/my-tashabbus/api/internal/modules/mfys"
	"github.com/my-tashabbus/api/internal/modules/users"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{service: service}
}

func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	current, mfyID, ok := currentUserAndID(w, r, "mfyID")
	if !ok {
		return
	}
	var input CreateStreetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	input.MFYID = mfyID
	street, err := h.service.Create(r.Context(), current, input)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusCreated, street)
}

func (h Handler) ListByMFY(w http.ResponseWriter, r *http.Request) {
	current, mfyID, ok := currentUserAndID(w, r, "mfyID")
	if !ok {
		return
	}
	limit := clamp(parseIntDefault(r.URL.Query().Get("limit"), 50), 1, 200)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	streets, err := h.service.ListByMFY(r.Context(), current, mfyID, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Paginated(w, http.StatusOK, streets, limit, offset)
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	street, err := h.service.Get(r.Context(), current, id)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusOK, street)
}

func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	var input UpdateStreetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	street, err := h.service.Update(r.Context(), current, id, input)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusOK, street)
}

func (h Handler) AssignLeader(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	var input struct {
		UserID     uuid.UUID `json:"user_id"`
		TelegramID *int64    `json:"telegram_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON request body")
		return
	}
	var assignment StreetLeaderAssignment
	var err error
	if input.TelegramID != nil {
		assignment, err = h.service.AssignLeaderByTelegramID(r.Context(), current, id, *input.TelegramID)
	} else {
		assignment, err = h.service.AssignLeader(r.Context(), current, id, input.UserID)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusOK, assignment)
}

func (h Handler) GetLeader(w http.ResponseWriter, r *http.Request) {
	current, id, ok := currentUserAndID(w, r, "id")
	if !ok {
		return
	}
	leader, err := h.service.GetLeader(r.Context(), current, id)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Data(w, http.StatusOK, leader)
}

func (h Handler) MyStreets(w http.ResponseWriter, r *http.Request) {
	current, ok := currentUser(w, r)
	if !ok {
		return
	}
	limit := clamp(parseIntDefault(r.URL.Query().Get("limit"), 50), 1, 200)
	offset := parseIntDefault(r.URL.Query().Get("offset"), 0)
	streets, err := h.service.MyStreets(r.Context(), current, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	response.Paginated(w, http.StatusOK, streets, limit, offset)
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

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrForbidden), errors.Is(err, mfys.ErrForbidden):
		response.ErrorCode(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource")
	case errors.Is(err, ErrStreetNotFound), errors.Is(err, mfys.ErrMFYNotFound), errors.Is(err, users.ErrUserNotFound):
		response.ErrorCode(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, ErrDuplicateStreetName):
		response.ErrorCode(w, http.StatusConflict, "CONFLICT", err.Error())
	case errors.Is(err, ErrNameRequired), errors.Is(err, ErrPlannedHouseholdsCount), errors.Is(err, ErrInvalidPagination), errors.Is(err, ErrStreetLeaderRoleRequired), errors.Is(err, ErrStreetLeaderWrongMFY):
		response.ErrorCode(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		response.ErrorCode(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
