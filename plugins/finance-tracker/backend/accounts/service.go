package accounts

import (
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Service contains the business logic for account operations.
type Service struct {
	repo *Repository
}

// NewService creates a Service backed by the given Repository.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListActive returns all active accounts with their computed balances.
func (s *Service) ListActive() ([]AccountWithBalance, error) {
	accounts, err := s.repo.ListActive()
	if err != nil {
		return nil, err
	}

	result := make([]AccountWithBalance, 0, len(accounts))
	for _, account := range accounts {
		balance, err := s.repo.CalculateBalance(account.ID)
		if err != nil {
			return nil, err
		}
		awb := AccountWithBalance{
			Account: account,
			Balance: balance,
		}
		awb.EstimatedInterest = estimateInterest(&account, balance)
		result = append(result, awb)
	}

	return result, nil
}

// Create validates input and creates a new account.
func (s *Service) Create(input *CreateAccountInput) (int64, *shared.AppError) {
	if appErr := validateCreateInput(input); appErr != nil {
		return 0, appErr
	}

	id, err := s.repo.Create(input)
	if err != nil {
		return 0, shared.NewAppError("INTERNAL", "failed to create account", 500)
	}
	return id, nil
}

// Update validates input and modifies an existing account.
func (s *Service) Update(id int64, input *UpdateAccountInput) *shared.AppError {
	if _, appErr := s.repo.GetByID(id); appErr != nil {
		return appErr
	}

	if appErr := validateUpdateInput(input); appErr != nil {
		return appErr
	}

	if err := s.repo.Update(id, input); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return appErr
		}
		return shared.NewAppError("INTERNAL", "failed to update account", 500)
	}
	return nil
}

// Archive soft-deletes an account by setting is_archived to 1.
func (s *Service) Archive(id int64) *shared.AppError {
	if err := s.repo.Archive(id); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return appErr
		}
		return shared.NewAppError("INTERNAL", "failed to archive account", 500)
	}
	return nil
}

// GetBalance returns the computed balance for a specific account, including
// estimated annual interest for savings accounts with interest_rate set.
func (s *Service) GetBalance(id int64) (*AccountWithBalance, *shared.AppError) {
	account, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}

	balance, err := s.repo.CalculateBalance(id)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to calculate balance", 500)
	}

	awb := &AccountWithBalance{
		Account: *account,
		Balance: balance,
	}
	awb.EstimatedInterest = estimateInterest(account, balance)

	return awb, nil
}

// estimateInterest calculates estimated annual interest for savings accounts.
// Returns nil if the account is not savings or has no interest_rate.
func estimateInterest(account *Account, balance float64) *float64 {
	if account.Type != "savings" || account.InterestRate == nil {
		return nil
	}
	interest := balance * (*account.InterestRate / 100)
	return &interest
}

// validateCreateInput checks that all required fields are present and valid.
func validateCreateInput(input *CreateAccountInput) *shared.AppError {
	if input.Name == "" {
		return shared.NewValidationError("name is required")
	}
	if !IsValidAccountType(input.Type) {
		return shared.NewValidationError("type must be one of: checking, savings, cash, investment")
	}
	if input.InterestRate != nil && input.Type != "savings" {
		return shared.NewValidationError("interest_rate is only allowed for savings accounts")
	}
	return nil
}

// validateUpdateInput checks that all required fields are present and valid.
func validateUpdateInput(input *UpdateAccountInput) *shared.AppError {
	if input.Name == "" {
		return shared.NewValidationError("name is required")
	}
	if !IsValidAccountType(input.Type) {
		return shared.NewValidationError("type must be one of: checking, savings, cash, investment")
	}
	if input.InterestRate != nil && input.Type != "savings" {
		return shared.NewValidationError("interest_rate is only allowed for savings accounts")
	}
	return nil
}
