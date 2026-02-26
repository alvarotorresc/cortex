package budgets

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Handler routes budget-related API requests to the appropriate service method.
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
	case req.Method == "GET" && req.Path == "/budgets":
		return h.list(req)
	case req.Method == "POST" && req.Path == "/budgets":
		return h.create(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/budgets/"):
		return h.update(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/budgets/"):
		return h.delete(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

func (h *Handler) list(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	month := req.Query["month"]
	if month == "" {
		return shared.JSONError(shared.NewValidationError("month query parameter is required (YYYY-MM)"))
	}

	budgets, appErr := h.service.List(month)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, budgets)
}

func (h *Handler) create(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input CreateBudgetInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	budget, appErr := h.service.Create(&input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(201, budget)
}

func (h *Handler) update(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	var input UpdateBudgetInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	budget, appErr := h.service.Update(id, &input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, budget)
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
