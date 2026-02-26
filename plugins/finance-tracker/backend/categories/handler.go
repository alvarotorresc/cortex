package categories

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Handler routes category-related API requests to the appropriate method.
type Handler struct {
	repo *Repository
}

// NewHandler creates a Handler with the repository wired to the given database.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: NewRepository(db)}
}

// Handle dispatches the request to the correct handler based on method and path.
// The reorder route must come before the generic PUT /categories/{id} route,
// otherwise "/categories/reorder" would match the update pattern.
func (h *Handler) Handle(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	case req.Method == "GET" && req.Path == "/categories":
		return h.list(req)
	case req.Method == "POST" && req.Path == "/categories":
		return h.create(req)
	case req.Method == "PUT" && req.Path == "/categories/reorder":
		return h.reorder(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/categories/"):
		return h.update(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/categories/"):
		return h.delete(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

func (h *Handler) list(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	typeFilter := ""
	if req.Query != nil {
		typeFilter = req.Query["type"]
	}

	categories, err := h.repo.List(typeFilter)
	if err != nil {
		return nil, err
	}
	return shared.JSONSuccess(200, categories)
}

func (h *Handler) create(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input CreateCategoryInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	if appErr := validateCreateInput(&input); appErr != nil {
		return shared.JSONError(appErr)
	}

	// Check for duplicate name (case-insensitive).
	exists, err := h.repo.ExistsByName(input.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return shared.JSONError(shared.NewConflictError(
			"category '" + input.Name + "' already exists",
		))
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

	var input UpdateCategoryInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	if appErr := validateUpdateInput(&input); appErr != nil {
		return shared.JSONError(appErr)
	}

	// Check for duplicate name excluding current category.
	exists, err := h.repo.ExistsByNameExcluding(input.Name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return shared.JSONError(shared.NewConflictError(
			"category '" + input.Name + "' already exists",
		))
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

	// Look up the category to get its name for the transactions check.
	category, appErr := h.repo.GetByID(id)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	// Check if any transactions reference this category.
	hasTransactions, err := h.repo.HasTransactions(category.Name)
	if err != nil {
		return nil, err
	}
	if hasTransactions {
		return shared.JSONError(shared.NewConflictError(
			"cannot delete category '" + category.Name + "': transactions reference it",
		))
	}

	if err := h.repo.Delete(id); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return shared.JSONError(appErr)
		}
		return nil, err
	}

	return shared.JSONSuccess(200, map[string]interface{}{"deleted": id})
}

func (h *Handler) reorder(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var items []ReorderItem
	if err := json.Unmarshal(req.Body, &items); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body: expected array of {id, sort_order}"))
	}

	if len(items) == 0 {
		return shared.JSONError(shared.NewValidationError("reorder list cannot be empty"))
	}

	if err := h.repo.Reorder(items); err != nil {
		return nil, err
	}

	return shared.JSONSuccess(200, map[string]interface{}{"reordered": len(items)})
}

// validateCreateInput checks that all required fields are present and valid.
func validateCreateInput(input *CreateCategoryInput) *shared.AppError {
	if strings.TrimSpace(input.Name) == "" {
		return shared.NewValidationError("name is required")
	}
	if !IsValidCategoryType(input.Type) {
		return shared.NewValidationError("type must be one of: income, expense, both")
	}
	return nil
}

// validateUpdateInput checks that all required fields are present and valid.
func validateUpdateInput(input *UpdateCategoryInput) *shared.AppError {
	if strings.TrimSpace(input.Name) == "" {
		return shared.NewValidationError("name is required")
	}
	if !IsValidCategoryType(input.Type) {
		return shared.NewValidationError("type must be one of: income, expense, both")
	}
	return nil
}
