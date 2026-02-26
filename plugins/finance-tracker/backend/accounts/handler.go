package accounts

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Handler routes account-related API requests to the appropriate service method.
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
	case req.Method == "GET" && req.Path == "/accounts":
		return h.list()
	case req.Method == "POST" && req.Path == "/accounts":
		return h.create(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/accounts/"):
		return h.update(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/accounts/"):
		return h.archive(req)
	case req.Method == "GET" && strings.HasSuffix(req.Path, "/balance"):
		return h.getBalance(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

func (h *Handler) list() (*sdk.APIResponse, error) {
	accounts, err := h.service.ListActive()
	if err != nil {
		return nil, err
	}
	return shared.JSONSuccess(200, accounts)
}

func (h *Handler) create(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input CreateAccountInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	if input.Currency == "" {
		input.Currency = "EUR"
	}

	id, appErr := h.service.Create(input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(201, map[string]interface{}{"id": id})
}

func (h *Handler) update(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	var input UpdateAccountInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	if input.Currency == "" {
		input.Currency = "EUR"
	}

	if appErr := h.service.Update(id, input); appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, map[string]interface{}{"updated": id})
}

func (h *Handler) archive(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	if appErr := h.service.Archive(id); appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, map[string]interface{}{"archived": id})
}

func (h *Handler) getBalance(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	// Path: /accounts/{id}/balance — extract id from second segment.
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	result, appErr := h.service.GetBalance(id)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, result)
}
