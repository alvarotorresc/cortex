package budgets

import (
	"math"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Service contains the business logic for budget operations.
type Service struct {
	repo *Repository
}

// NewService creates a Service backed by the given Repository.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns all budgets for the given month enriched with spending progress.
// It includes recurring budgets (month is NULL/empty) alongside month-specific ones.
func (s *Service) List(month string) ([]BudgetWithProgress, *shared.AppError) {
	if !IsValidMonth(month) {
		return nil, shared.NewValidationError("month must be in YYYY-MM format")
	}

	budgets, err := s.repo.List(month)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to list budgets", 500)
	}

	result := make([]BudgetWithProgress, 0, len(budgets))
	for i := range budgets {
		b := &budgets[i]

		// Determine query month: use the budget's month if set, otherwise the requested month.
		queryMonth := month
		if b.Month != "" {
			queryMonth = b.Month
		}

		var spent float64
		if b.Category == "" {
			// Global budget: sum all expenses for the month.
			spent, err = s.repo.CalculateSpentGlobal(queryMonth)
		} else {
			// Category budget: sum only that category's expenses.
			spent, err = s.repo.CalculateSpentByCategory(queryMonth, b.Category)
		}
		if err != nil {
			return nil, shared.NewAppError("INTERNAL", "failed to calculate spending", 500)
		}

		remaining := b.Amount - spent
		var percentage float64
		if b.Amount > 0 {
			percentage = math.Round((spent/b.Amount)*10000) / 100
		}

		result = append(result, BudgetWithProgress{
			Budget:     *b,
			Spent:      spent,
			Remaining:  remaining,
			Percentage: percentage,
		})
	}

	return result, nil
}

// Create validates input and inserts a new budget.
func (s *Service) Create(input *CreateBudgetInput) (*Budget, *shared.AppError) {
	if appErr := validateCreateInput(input); appErr != nil {
		return nil, appErr
	}

	id, err := s.repo.Create(input)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to create budget", 500)
	}

	budget, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return budget, nil
}

// Update validates input and modifies an existing budget.
func (s *Service) Update(id int64, input *UpdateBudgetInput) (*Budget, *shared.AppError) {
	// Verify budget exists.
	if _, appErr := s.repo.GetByID(id); appErr != nil {
		return nil, appErr
	}

	if appErr := validateUpdateInput(input); appErr != nil {
		return nil, appErr
	}

	if err := s.repo.Update(id, input); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return nil, appErr
		}
		return nil, shared.NewAppError("INTERNAL", "failed to update budget", 500)
	}

	budget, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return budget, nil
}

// Delete removes a budget permanently.
func (s *Service) Delete(id int64) *shared.AppError {
	if err := s.repo.Delete(id); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return appErr
		}
		return shared.NewAppError("INTERNAL", "failed to delete budget", 500)
	}
	return nil
}

// validateCreateInput checks that all required fields are present and valid.
func validateCreateInput(input *CreateBudgetInput) *shared.AppError {
	if input.Amount <= 0 {
		return shared.NewValidationError("amount must be greater than 0")
	}
	if input.Month != "" && !IsValidMonth(input.Month) {
		return shared.NewValidationError("month must be in YYYY-MM format")
	}
	return nil
}

// validateUpdateInput checks that all required fields are present and valid.
func validateUpdateInput(input *UpdateBudgetInput) *shared.AppError {
	if input.Amount <= 0 {
		return shared.NewValidationError("amount must be greater than 0")
	}
	if input.Month != "" && !IsValidMonth(input.Month) {
		return shared.NewValidationError("month must be in YYYY-MM format")
	}
	return nil
}
