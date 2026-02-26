package transactions

import (
	"fmt"
	"strings"
	"time"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Service contains the business logic for transaction operations.
type Service struct {
	repo *Repository
}

// NewService creates a Service backed by the given Repository.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns all transactions matching the given filter.
func (s *Service) List(filter *TransactionFilter) ([]Transaction, error) {
	return s.repo.List(filter)
}

// Create validates input, applies defaults, inserts the transaction, and links tags.
func (s *Service) Create(input *CreateTransactionInput) (*Transaction, *shared.AppError) {
	if appErr := validateCreateInput(input); appErr != nil {
		return nil, appErr
	}

	// Default account_id to 1 (backward compatibility with v1).
	if input.AccountID == nil {
		defaultID := int64(1)
		input.AccountID = &defaultID
	}

	// Default date to today.
	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}

	// Validate account exists.
	if appErr := s.validateAccountExists(*input.AccountID); appErr != nil {
		return nil, appErr
	}

	// For transfers, validate dest_account_id.
	if input.Type == "transfer" {
		if appErr := s.validateAccountExists(*input.DestAccountID); appErr != nil {
			return nil, appErr
		}
	}

	id, err := s.repo.Create(input)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to create transaction", 500)
	}

	tx, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return tx, nil
}

// Update validates input, modifies the transaction, and updates tag links.
func (s *Service) Update(id int64, input *UpdateTransactionInput) (*Transaction, *shared.AppError) {
	// Verify transaction exists.
	if _, appErr := s.repo.GetByID(id); appErr != nil {
		return nil, appErr
	}

	if appErr := validateUpdateInput(input); appErr != nil {
		return nil, appErr
	}

	// Default account_id to 1.
	if input.AccountID == nil {
		defaultID := int64(1)
		input.AccountID = &defaultID
	}

	// Default date to today.
	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}

	// Validate account exists.
	if appErr := s.validateAccountExists(*input.AccountID); appErr != nil {
		return nil, appErr
	}

	// For transfers, validate dest_account_id.
	if input.Type == "transfer" {
		if appErr := s.validateAccountExists(*input.DestAccountID); appErr != nil {
			return nil, appErr
		}
	}

	if err := s.repo.Update(id, input); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return nil, appErr
		}
		return nil, shared.NewAppError("INTERNAL", "failed to update transaction", 500)
	}

	tx, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return tx, nil
}

// Delete removes a transaction by its ID.
func (s *Service) Delete(id int64) *shared.AppError {
	if err := s.repo.Delete(id); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return appErr
		}
		return shared.NewAppError("INTERNAL", "failed to delete transaction", 500)
	}
	return nil
}

// validateAccountExists checks that an account with the given ID exists.
func (s *Service) validateAccountExists(accountID int64) *shared.AppError {
	exists, err := s.repo.AccountExists(accountID)
	if err != nil {
		return shared.NewAppError("INTERNAL", "failed to check account", 500)
	}
	if !exists {
		return shared.NewValidationError(fmt.Sprintf("account %d not found", accountID))
	}
	return nil
}

// validateCreateInput checks that all required fields are present and valid.
func validateCreateInput(input *CreateTransactionInput) *shared.AppError {
	if input.Amount <= 0 {
		return shared.NewValidationError("amount must be greater than 0")
	}
	if !IsValidTransactionType(input.Type) {
		return shared.NewValidationError("type must be 'income', 'expense', or 'transfer'")
	}
	if input.Type == "transfer" {
		if input.DestAccountID == nil {
			return shared.NewValidationError("dest_account_id is required for transfers")
		}
		if input.AccountID != nil && *input.AccountID == *input.DestAccountID {
			return shared.NewValidationError("dest_account_id must differ from account_id")
		}
	}
	if input.Type != "transfer" && strings.TrimSpace(input.Category) == "" {
		return shared.NewValidationError("category is required")
	}
	return nil
}

// validateUpdateInput checks that all required fields are present and valid.
func validateUpdateInput(input *UpdateTransactionInput) *shared.AppError {
	if input.Amount <= 0 {
		return shared.NewValidationError("amount must be greater than 0")
	}
	if !IsValidTransactionType(input.Type) {
		return shared.NewValidationError("type must be 'income', 'expense', or 'transfer'")
	}
	if input.Type == "transfer" {
		if input.DestAccountID == nil {
			return shared.NewValidationError("dest_account_id is required for transfers")
		}
		if input.AccountID != nil && *input.AccountID == *input.DestAccountID {
			return shared.NewValidationError("dest_account_id must differ from account_id")
		}
	}
	if input.Type != "transfer" && strings.TrimSpace(input.Category) == "" {
		return shared.NewValidationError("category is required")
	}
	return nil
}
