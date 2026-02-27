package reports

import (
	"database/sql"
	"time"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Handler routes report-related API requests to the service layer.
type Handler struct {
	service *Service
}

// NewHandler creates a Handler wired to the reports service.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{service: NewService(db)}
}

// Handle dispatches the request to the correct report handler.
func (h *Handler) Handle(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	case req.Method == "GET" && req.Path == "/reports/summary":
		return h.summary(req)
	case req.Method == "GET" && req.Path == "/reports/trends":
		return h.trends(req)
	case req.Method == "GET" && req.Path == "/reports/categories":
		return h.categories(req)
	case req.Method == "GET" && req.Path == "/reports/net-worth":
		return h.netWorth()
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

func (h *Handler) summary(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	month := req.Query["month"]
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	if err := validateMonth(month); err != nil {
		return shared.JSONError(err)
	}

	summary, appErr := h.service.Summary(month)
	if appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(200, summary)
}

func (h *Handler) trends(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	from := req.Query["from"]
	to := req.Query["to"]

	if from == "" || to == "" {
		return shared.JSONError(shared.NewValidationError("'from' and 'to' query parameters are required"))
	}

	if err := validateMonth(from); err != nil {
		return shared.JSONError(err)
	}
	if err := validateMonth(to); err != nil {
		return shared.JSONError(err)
	}

	trends, appErr := h.service.Trends(from, to)
	if appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(200, trends)
}

func (h *Handler) categories(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	month := req.Query["month"]
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	if err := validateMonth(month); err != nil {
		return shared.JSONError(err)
	}

	comparisons, appErr := h.service.Categories(month)
	if appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(200, comparisons)
}

func (h *Handler) netWorth() (*sdk.APIResponse, error) {
	nw, appErr := h.service.NetWorthReport()
	if appErr != nil {
		return shared.JSONError(appErr)
	}
	return shared.JSONSuccess(200, nw)
}

// validateMonth checks that a string is in YYYY-MM format.
func validateMonth(month string) *shared.AppError {
	if _, err := time.Parse("2006-01", month); err != nil {
		return shared.NewValidationError("month must be in YYYY-MM format")
	}
	return nil
}
