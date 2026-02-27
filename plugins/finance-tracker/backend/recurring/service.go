package recurring

import (
	"fmt"
	"strings"
	"time"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Service contains the business logic for recurring rule operations.
type Service struct {
	repo *Repository
}

// NewService creates a Service backed by the given Repository.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// List returns all recurring rules.
func (s *Service) List() ([]Rule, error) {
	return s.repo.List()
}

// Create validates input, applies defaults, and inserts a recurring rule.
func (s *Service) Create(input *CreateRuleInput) (*Rule, *shared.AppError) {
	if appErr := validateCreateInput(input); appErr != nil {
		return nil, appErr
	}

	// Default account_id to 1.
	if input.AccountID == nil {
		defaultID := int64(1)
		input.AccountID = &defaultID
	}

	// Validate account exists.
	if appErr := s.validateAccountExists(*input.AccountID); appErr != nil {
		return nil, appErr
	}

	if input.Type == "transfer" && input.DestAccountID != nil {
		if appErr := s.validateAccountExists(*input.DestAccountID); appErr != nil {
			return nil, appErr
		}
	}

	id, err := s.repo.Create(input)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", "failed to create recurring rule", 500)
	}

	rule, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return rule, nil
}

// Update validates input and modifies an existing recurring rule.
func (s *Service) Update(id int64, input *UpdateRuleInput) (*Rule, *shared.AppError) {
	// Verify rule exists.
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

	if appErr := s.validateAccountExists(*input.AccountID); appErr != nil {
		return nil, appErr
	}

	if input.Type == "transfer" && input.DestAccountID != nil {
		if appErr := s.validateAccountExists(*input.DestAccountID); appErr != nil {
			return nil, appErr
		}
	}

	if err := s.repo.Update(id, input); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return nil, appErr
		}
		return nil, shared.NewAppError("INTERNAL", "failed to update recurring rule", 500)
	}

	rule, appErr := s.repo.GetByID(id)
	if appErr != nil {
		return nil, appErr
	}
	return rule, nil
}

// Deactivate sets is_active=0 for a rule (soft delete).
func (s *Service) Deactivate(id int64) *shared.AppError {
	if err := s.repo.Deactivate(id); err != nil {
		if appErr, ok := err.(*shared.AppError); ok {
			return appErr
		}
		return shared.NewAppError("INTERNAL", "failed to deactivate recurring rule", 500)
	}
	return nil
}

// Generate creates pending transaction instances for all active rules up to the given date.
// It is idempotent: calling it twice will not create duplicate transactions.
func (s *Service) Generate(today time.Time) (*GenerateResult, *shared.AppError) {
	todayStr := today.Format("2006-01-02")

	rules, err := s.repo.ListActiveRules(todayStr)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("listing active rules: %v", err), 500)
	}

	totalGenerated := 0

	for i := range rules {
		rule := &rules[i]
		count, appErr := s.generateForRule(rule, today)
		if appErr != nil {
			return nil, appErr
		}
		totalGenerated += count
	}

	return &GenerateResult{Generated: totalGenerated}, nil
}

// generateForRule calculates all pending dates for a single rule and inserts transactions.
func (s *Service) generateForRule(rule *Rule, today time.Time) (int, *shared.AppError) {
	// Determine the starting point.
	var startFrom time.Time
	if rule.LastGenerated != "" {
		parsed, err := time.Parse("2006-01-02", rule.LastGenerated)
		if err != nil {
			return 0, shared.NewAppError("INTERNAL",
				fmt.Sprintf("parsing last_generated for rule %d: %v", rule.ID, err), 500)
		}
		// Start from the day after last generated.
		startFrom = parsed.AddDate(0, 0, 1)
	} else {
		parsed, err := time.Parse("2006-01-02", rule.StartDate)
		if err != nil {
			return 0, shared.NewAppError("INTERNAL",
				fmt.Sprintf("parsing start_date for rule %d: %v", rule.ID, err), 500)
		}
		startFrom = parsed
	}

	// Determine end boundary.
	endBoundary := today
	if rule.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", rule.EndDate)
		if err != nil {
			return 0, shared.NewAppError("INTERNAL",
				fmt.Sprintf("parsing end_date for rule %d: %v", rule.ID, err), 500)
		}
		if endDate.Before(endBoundary) {
			endBoundary = endDate
		}
	}

	// Calculate all pending dates.
	dates := calculateDates(rule, startFrom, endBoundary)

	generated := 0
	var lastDate string

	for _, date := range dates {
		dateStr := date.Format("2006-01-02")

		// Idempotency check: skip if already generated.
		exists, err := s.repo.TransactionExistsForDate(rule.ID, dateStr)
		if err != nil {
			return 0, shared.NewAppError("INTERNAL",
				fmt.Sprintf("checking existing transaction: %v", err), 500)
		}
		if exists {
			lastDate = dateStr
			continue
		}

		err = s.repo.InsertGeneratedTransaction(
			rule.ID, rule.Amount, rule.Type, rule.AccountID,
			rule.DestAccountID, rule.Category, rule.Description, dateStr,
		)
		if err != nil {
			return 0, shared.NewAppError("INTERNAL",
				fmt.Sprintf("inserting generated transaction: %v", err), 500)
		}
		generated++
		lastDate = dateStr
	}

	// Update last_generated to the last date we processed.
	if lastDate != "" {
		if err := s.repo.UpdateLastGenerated(rule.ID, lastDate); err != nil {
			return 0, shared.NewAppError("INTERNAL",
				fmt.Sprintf("updating last_generated: %v", err), 500)
		}
	}

	// If end_date is set and has passed, deactivate the rule.
	if rule.EndDate != "" {
		endDate, _ := time.Parse("2006-01-02", rule.EndDate)
		if !endDate.After(today) {
			if err := s.repo.DeactivateRule(rule.ID); err != nil {
				return 0, shared.NewAppError("INTERNAL",
					fmt.Sprintf("deactivating expired rule: %v", err), 500)
			}
		}
	}

	return generated, nil
}

// calculateDates computes all occurrence dates for a rule between startFrom and endBoundary (inclusive).
func calculateDates(rule *Rule, startFrom time.Time, endBoundary time.Time) []time.Time {
	var dates []time.Time

	switch rule.Frequency {
	case "monthly":
		dates = calculateMonthlyDates(rule, startFrom, endBoundary)
	case "weekly":
		dates = calculateWeeklyDates(rule, startFrom, endBoundary, 7)
	case "biweekly":
		dates = calculateWeeklyDates(rule, startFrom, endBoundary, 14)
	case "yearly":
		dates = calculateYearlyDates(rule, startFrom, endBoundary)
	}

	return dates
}

// calculateMonthlyDates generates monthly dates using day_of_month.
// Handles edge case: day_of_month=31 in a 30-day month uses the last day of the month.
func calculateMonthlyDates(rule *Rule, startFrom time.Time, endBoundary time.Time) []time.Time {
	if rule.DayOfMonth == nil {
		return nil
	}

	dayOfMonth := *rule.DayOfMonth
	var dates []time.Time

	// Start from the month of startFrom.
	current := time.Date(startFrom.Year(), startFrom.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !current.After(endBoundary) {
		date := clampDayOfMonth(current.Year(), current.Month(), dayOfMonth)

		if !date.Before(startFrom) && !date.After(endBoundary) {
			dates = append(dates, date)
		}

		// Advance to next month.
		current = current.AddDate(0, 1, 0)
	}

	return dates
}

// calculateWeeklyDates generates weekly or biweekly dates using day_of_week.
func calculateWeeklyDates(rule *Rule, startFrom time.Time, endBoundary time.Time, intervalDays int) []time.Time {
	if rule.DayOfWeek == nil {
		return nil
	}

	targetDay := time.Weekday(*rule.DayOfWeek)
	var dates []time.Time

	// Find the first occurrence of the target weekday on or after startFrom.
	current := startFrom
	for current.Weekday() != targetDay {
		current = current.AddDate(0, 0, 1)
	}

	for !current.After(endBoundary) {
		dates = append(dates, current)
		current = current.AddDate(0, 0, intervalDays)
	}

	return dates
}

// calculateYearlyDates generates yearly dates using day_of_month + month_of_year.
func calculateYearlyDates(rule *Rule, startFrom time.Time, endBoundary time.Time) []time.Time {
	if rule.DayOfMonth == nil || rule.MonthOfYear == nil {
		return nil
	}

	dayOfMonth := *rule.DayOfMonth
	monthOfYear := time.Month(*rule.MonthOfYear)
	var dates []time.Time

	for year := startFrom.Year(); ; year++ {
		date := clampDayOfMonth(year, monthOfYear, dayOfMonth)

		if date.After(endBoundary) {
			break
		}

		if !date.Before(startFrom) {
			dates = append(dates, date)
		}
	}

	return dates
}

// clampDayOfMonth returns a date for the given year/month/day, clamping the day to the
// last day of the month if it exceeds the month's length (e.g. day=31 in April returns April 30).
func clampDayOfMonth(year int, month time.Month, day int) time.Time {
	// Get the last day of the month by going to the first of next month and subtracting a day.
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
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
func validateCreateInput(input *CreateRuleInput) *shared.AppError {
	if input.Amount <= 0 {
		return shared.NewValidationError("amount must be greater than 0")
	}
	if !IsValidType(input.Type) {
		return shared.NewValidationError("type must be 'income', 'expense', or 'transfer'")
	}
	if !IsValidFrequency(input.Frequency) {
		return shared.NewValidationError("frequency must be 'weekly', 'biweekly', 'monthly', or 'yearly'")
	}
	if strings.TrimSpace(input.StartDate) == "" {
		return shared.NewValidationError("start_date is required")
	}
	if _, err := time.Parse("2006-01-02", input.StartDate); err != nil {
		return shared.NewValidationError("start_date must be in YYYY-MM-DD format")
	}
	if input.EndDate != "" {
		if _, err := time.Parse("2006-01-02", input.EndDate); err != nil {
			return shared.NewValidationError("end_date must be in YYYY-MM-DD format")
		}
	}
	if input.Type != "transfer" && strings.TrimSpace(input.Category) == "" {
		return shared.NewValidationError("category is required")
	}
	if input.Type == "transfer" {
		if input.DestAccountID == nil {
			return shared.NewValidationError("dest_account_id is required for transfers")
		}
		if input.AccountID != nil && *input.AccountID == *input.DestAccountID {
			return shared.NewValidationError("dest_account_id must differ from account_id")
		}
	}

	// Frequency-specific field validation.
	if input.Frequency == "monthly" || input.Frequency == "yearly" {
		if input.DayOfMonth == nil {
			return shared.NewValidationError("day_of_month is required for monthly/yearly frequency")
		}
		if *input.DayOfMonth < 1 || *input.DayOfMonth > 31 {
			return shared.NewValidationError("day_of_month must be between 1 and 31")
		}
	}
	if input.Frequency == "weekly" || input.Frequency == "biweekly" {
		if input.DayOfWeek == nil {
			return shared.NewValidationError("day_of_week is required for weekly/biweekly frequency")
		}
		if *input.DayOfWeek < 0 || *input.DayOfWeek > 6 {
			return shared.NewValidationError("day_of_week must be between 0 and 6")
		}
	}
	if input.Frequency == "yearly" {
		if input.MonthOfYear == nil {
			return shared.NewValidationError("month_of_year is required for yearly frequency")
		}
		if *input.MonthOfYear < 1 || *input.MonthOfYear > 12 {
			return shared.NewValidationError("month_of_year must be between 1 and 12")
		}
	}

	return nil
}

// validateUpdateInput checks that all required fields are present and valid.
func validateUpdateInput(input *UpdateRuleInput) *shared.AppError {
	if input.Amount <= 0 {
		return shared.NewValidationError("amount must be greater than 0")
	}
	if !IsValidType(input.Type) {
		return shared.NewValidationError("type must be 'income', 'expense', or 'transfer'")
	}
	if !IsValidFrequency(input.Frequency) {
		return shared.NewValidationError("frequency must be 'weekly', 'biweekly', 'monthly', or 'yearly'")
	}
	if strings.TrimSpace(input.StartDate) == "" {
		return shared.NewValidationError("start_date is required")
	}
	if _, err := time.Parse("2006-01-02", input.StartDate); err != nil {
		return shared.NewValidationError("start_date must be in YYYY-MM-DD format")
	}
	if input.EndDate != "" {
		if _, err := time.Parse("2006-01-02", input.EndDate); err != nil {
			return shared.NewValidationError("end_date must be in YYYY-MM-DD format")
		}
	}
	if input.Type != "transfer" && strings.TrimSpace(input.Category) == "" {
		return shared.NewValidationError("category is required")
	}
	if input.Type == "transfer" {
		if input.DestAccountID == nil {
			return shared.NewValidationError("dest_account_id is required for transfers")
		}
		if input.AccountID != nil && *input.AccountID == *input.DestAccountID {
			return shared.NewValidationError("dest_account_id must differ from account_id")
		}
	}

	if input.Frequency == "monthly" || input.Frequency == "yearly" {
		if input.DayOfMonth == nil {
			return shared.NewValidationError("day_of_month is required for monthly/yearly frequency")
		}
		if *input.DayOfMonth < 1 || *input.DayOfMonth > 31 {
			return shared.NewValidationError("day_of_month must be between 1 and 31")
		}
	}
	if input.Frequency == "weekly" || input.Frequency == "biweekly" {
		if input.DayOfWeek == nil {
			return shared.NewValidationError("day_of_week is required for weekly/biweekly frequency")
		}
		if *input.DayOfWeek < 0 || *input.DayOfWeek > 6 {
			return shared.NewValidationError("day_of_week must be between 0 and 6")
		}
	}
	if input.Frequency == "yearly" {
		if input.MonthOfYear == nil {
			return shared.NewValidationError("month_of_year is required for yearly frequency")
		}
		if *input.MonthOfYear < 1 || *input.MonthOfYear > 12 {
			return shared.NewValidationError("month_of_year must be between 1 and 12")
		}
	}

	return nil
}
