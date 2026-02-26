package goals

import (
	"strings"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Service contains the business logic for savings goal operations.
type Service struct {
	repo *Repository
}

// NewService creates a Service backed by the given Repository.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns all savings goals.
func (s *Service) List() ([]SavingsGoal, *shared.AppError) {
	goals, err := s.repo.List()
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to list goals", 500)
	}
	return goals, nil
}

// Create validates input and inserts a new savings goal.
func (s *Service) Create(input *CreateGoalInput) (*SavingsGoal, *shared.AppError) {
	if appErr := validateCreateInput(input); appErr != nil {
		return nil, appErr
	}

	id, err := s.repo.Create(input)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to create goal", 500)
	}

	goal, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return goal, nil
}

// Update validates input and modifies an existing savings goal.
func (s *Service) Update(id int64, input *UpdateGoalInput) (*SavingsGoal, *shared.AppError) {
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
		return nil, shared.NewAppError("INTERNAL", "failed to update goal", 500)
	}

	goal, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return goal, nil
}

// Delete removes a savings goal permanently.
func (s *Service) Delete(id int64) *shared.AppError {
	if err := s.repo.Delete(id); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return appErr
		}
		return shared.NewAppError("INTERNAL", "failed to delete goal", 500)
	}
	return nil
}

// Contribute adds an amount to a savings goal and auto-completes if target is reached.
func (s *Service) Contribute(id int64, input *ContributeInput) (*SavingsGoal, *shared.AppError) {
	if input.Amount <= 0 {
		return nil, shared.NewValidationError("amount must be greater than 0")
	}

	goal, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}

	newAmount := goal.CurrentAmount + input.Amount
	isCompleted := newAmount >= goal.TargetAmount

	if err := s.repo.UpdateAmountAndCompletion(id, newAmount, isCompleted); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return nil, appErr
		}
		return nil, shared.NewAppError("INTERNAL", "failed to contribute to goal", 500)
	}

	updated, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return updated, nil
}

// --- Validation ---

func validateCreateInput(input *CreateGoalInput) *shared.AppError {
	if strings.TrimSpace(input.Name) == "" {
		return shared.NewValidationError("name is required")
	}
	if input.TargetAmount <= 0 {
		return shared.NewValidationError("target_amount must be greater than 0")
	}
	return nil
}

func validateUpdateInput(input *UpdateGoalInput) *shared.AppError {
	if strings.TrimSpace(input.Name) == "" {
		return shared.NewValidationError("name is required")
	}
	if input.TargetAmount <= 0 {
		return shared.NewValidationError("target_amount must be greater than 0")
	}
	return nil
}
