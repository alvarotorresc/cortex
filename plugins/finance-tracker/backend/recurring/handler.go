package recurring

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Handler routes recurring-rule-related API requests to the appropriate service method.
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
	case req.Method == "GET" && req.Path == "/recurring":
		return h.list(req)
	case req.Method == "POST" && req.Path == "/recurring":
		return h.create(req)
	case req.Method == "POST" && req.Path == "/recurring/generate":
		return h.generate(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/recurring/"):
		return h.update(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/recurring/"):
		return h.delete(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

func (h *Handler) list(_ *sdk.APIRequest) (*sdk.APIResponse, error) {
	rules, err := h.service.List()
	if err != nil {
		return nil, err
	}
	return shared.JSONSuccess(200, rules)
}

func (h *Handler) create(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input CreateRuleInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	rule, appErr := h.service.Create(&input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(201, rule)
}

func (h *Handler) update(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	var input UpdateRuleInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	rule, appErr := h.service.Update(id, &input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, rule)
}

func (h *Handler) delete(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	if appErr := h.service.Deactivate(id); appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, map[string]interface{}{"deactivated": id})
}

func (h *Handler) generate(_ *sdk.APIRequest) (*sdk.APIResponse, error) {
	result, appErr := h.service.Generate(time.Now())
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, result)
}
