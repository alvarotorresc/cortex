package goals

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Handler routes savings-goal-related API requests to the appropriate service method.
type Handler struct {
	service *Service
}

// NewHandler creates a Handler with all layers wired together.
func NewHandler(db *sql.DB) *Handler {
	repo := NewRepository(db)
	svc := NewService(repo)
	return &Handler{service: svc}
}

// Handle dispatches the request to the correct handler based on method and path.
func (h *Handler) Handle(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	// Contribute must be matched before generic POST /goals.
	case req.Method == "POST" && strings.Contains(req.Path, "/contribute"):
		return h.contribute(req)
	case req.Method == "GET" && req.Path == "/goals":
		return h.list()
	case req.Method == "POST" && req.Path == "/goals":
		return h.create(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/goals/"):
		return h.update(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/goals/"):
		return h.delete(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

func (h *Handler) list() (*sdk.APIResponse, error) {
	goals, appErr := h.service.List()
	if appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(200, goals)
}

func (h *Handler) create(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input CreateGoalInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	goal, appErr := h.service.Create(&input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(201, goal)
}

func (h *Handler) update(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	var input UpdateGoalInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	goal, appErr := h.service.Update(id, &input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(200, goal)
}

func (h *Handler) delete(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	if appErr := h.service.Delete(id); appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(200, map[string]interface{}{"deleted": id})
}

func (h *Handler) contribute(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	var input ContributeInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	goal, appErr := h.service.Contribute(id, &input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(200, goal)
}
