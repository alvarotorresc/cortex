package tags

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Handler routes tag-related API requests to the appropriate method.
type Handler struct {
	repo *Repository
}

// NewHandler creates a Handler with the repository wired to the given database.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: NewRepository(db)}
}

// Handle dispatches the request to the correct handler based on method and path.
func (h *Handler) Handle(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	case req.Method == "GET" && req.Path == "/tags":
		return h.list()
	case req.Method == "POST" && req.Path == "/tags":
		return h.create(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/tags/"):
		return h.update(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/tags/"):
		return h.delete(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

func (h *Handler) list() (*sdk.APIResponse, error) {
	tags, err := h.repo.List()
	if err != nil {
		return nil, err
	}
	return shared.JSONSuccess(200, tags)
}

func (h *Handler) create(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input CreateTagInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	if appErr := validateCreateInput(&input); appErr != nil {
		return shared.JSONError(appErr)
	}

	id, err := h.repo.Create(&input)
	if err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return shared.JSONError(appErr)
		}
		return nil, err
	}

	return shared.JSONSuccess(201, map[string]interface{}{"id": id})
}

func (h *Handler) update(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	var input UpdateTagInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	if appErr := validateUpdateInput(&input); appErr != nil {
		return shared.JSONError(appErr)
	}

	if err := h.repo.Update(id, &input); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return shared.JSONError(appErr)
		}
		return nil, err
	}

	return shared.JSONSuccess(200, map[string]interface{}{"updated": id})
}

func (h *Handler) delete(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	if err := h.repo.Delete(id); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return shared.JSONError(appErr)
		}
		return nil, err
	}

	return shared.JSONSuccess(200, map[string]interface{}{"deleted": id})
}

// validateCreateInput checks that all required fields are present and valid.
func validateCreateInput(input *CreateTagInput) *shared.AppError {
	if strings.TrimSpace(input.Name) == "" {
		return shared.NewValidationError("name is required")
	}
	return nil
}

// validateUpdateInput checks that all required fields are present and valid.
func validateUpdateInput(input *UpdateTagInput) *shared.AppError {
	if strings.TrimSpace(input.Name) == "" {
		return shared.NewValidationError("name is required")
	}
	return nil
}
