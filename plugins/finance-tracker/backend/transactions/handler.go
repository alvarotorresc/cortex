package transactions

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Handler routes transaction-related API requests to the appropriate service method.
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
	case req.Method == "GET" && req.Path == "/transactions":
		return h.list(req)
	case req.Method == "POST" && req.Path == "/transactions":
		return h.create(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/transactions/"):
		return h.update(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/transactions/"):
		return h.delete(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

func (h *Handler) list(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	filter := &TransactionFilter{
		Month:    req.Query["month"],
		Account:  req.Query["account"],
		Category: req.Query["category"],
		Tag:      req.Query["tag"],
		Type:     req.Query["type"],
		Search:   req.Query["search"],
	}

	// Validate month format if provided.
	if filter.Month != "" {
		if _, err := time.Parse("2006-01", filter.Month); err != nil {
			return shared.JSONError(shared.NewValidationError("month must be in YYYY-MM format"))
		}
	}

	// Default month to current if no filters are provided.
	if filter.Month == "" && filter.Account == "" && filter.Category == "" &&
		filter.Tag == "" && filter.Type == "" && filter.Search == "" {
		filter.Month = time.Now().Format("2006-01")
	}

	transactions, err := h.service.List(filter)
	if err != nil {
		return nil, err
	}
	return shared.JSONSuccess(200, transactions)
}

func (h *Handler) create(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input CreateTransactionInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	tx, appErr := h.service.Create(&input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(201, tx)
}

func (h *Handler) update(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id, appErr := shared.ExtractIDFromPath(req.Path)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	var input UpdateTransactionInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return shared.JSONError(shared.NewValidationError("invalid JSON body"))
	}

	tx, appErr := h.service.Update(id, &input)
	if appErr != nil {
		return shared.JSONError(appErr)
	}

	return shared.JSONSuccess(200, tx)
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
