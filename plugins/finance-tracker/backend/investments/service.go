package investments

import (
	"strings"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Service contains the business logic for investment operations.
type Service struct {
	repo *Repository
}

// NewService creates a Service backed by the given Repository.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns all investments with calculated P&L fields.
func (s *Service) List() ([]InvestmentWithPnL, *shared.AppError) {
	investments, err := s.repo.List()
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to list investments", 500)
	}

	result := make([]InvestmentWithPnL, 0, len(investments))
	for _, inv := range investments {
		result = append(result, calculatePnL(inv))
	}
	return result, nil
}

// Create validates input and inserts a new investment.
func (s *Service) Create(input *CreateInvestmentInput) (*Investment, *shared.AppError) {
	if appErr := validateCreateInput(input); appErr != nil {
		return nil, appErr
	}

	if input.Currency == "" {
		input.Currency = "EUR"
	}

	id, err := s.repo.Create(input)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to create investment", 500)
	}

	inv, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return inv, nil
}

// Update validates input and modifies an existing investment.
func (s *Service) Update(id int64, input *UpdateInvestmentInput) (*InvestmentWithPnL, *shared.AppError) {
	if _, appErr := s.repo.GetByID(id); appErr != nil {
		return nil, appErr
	}

	if appErr := validateUpdateInput(input); appErr != nil {
		return nil, appErr
	}

	if input.Currency == "" {
		input.Currency = "EUR"
	}

	if err := s.repo.Update(id, input); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return nil, appErr
		}
		return nil, shared.NewAppError("INTERNAL", "failed to update investment", 500)
	}

	inv, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}

	withPnL := calculatePnL(*inv)
	return &withPnL, nil
}

// Delete removes an investment permanently.
func (s *Service) Delete(id int64) *shared.AppError {
	if err := s.repo.Delete(id); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return appErr
		}
		return shared.NewAppError("INTERNAL", "failed to delete investment", 500)
	}
	return nil
}

// calculatePnL computes P&L fields for an investment.
// Only calculates when units, avg_buy_price, AND current_price are all non-nil.
func calculatePnL(inv Investment) InvestmentWithPnL {
	result := InvestmentWithPnL{Investment: inv}

	if inv.Units == nil || inv.AvgBuyPrice == nil || inv.CurrentPrice == nil {
		return result
	}

	totalInvested := *inv.Units * *inv.AvgBuyPrice
	currentValue := *inv.Units * *inv.CurrentPrice
	pnl := currentValue - totalInvested

	result.TotalInvested = &totalInvested
	result.CurrentValue = &currentValue
	result.PnL = &pnl

	// Guard against division by zero.
	if totalInvested != 0 {
		pnlPct := (pnl / totalInvested) * 100
		result.PnLPercentage = &pnlPct
	}

	return result
}

// --- Validation ---

func validateCreateInput(input *CreateInvestmentInput) *shared.AppError {
	if strings.TrimSpace(input.Name) == "" {
		return shared.NewValidationError("name is required")
	}
	if !validInvestmentTypes[input.Type] {
		return shared.NewValidationError("type must be one of: crypto, etf, fund, stock, other")
	}
	return nil
}

func validateUpdateInput(input *UpdateInvestmentInput) *shared.AppError {
	if strings.TrimSpace(input.Name) == "" {
		return shared.NewValidationError("name is required")
	}
	if !validInvestmentTypes[input.Type] {
		return shared.NewValidationError("type must be one of: crypto, etf, fund, stock, other")
	}
	return nil
}
