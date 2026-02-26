package goals

// SavingsGoal represents a savings goal with progress tracking.
type SavingsGoal struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	TargetAmount  float64 `json:"target_amount"`
	CurrentAmount float64 `json:"current_amount"`
	TargetDate    *string `json:"target_date,omitempty"`
	Icon          string  `json:"icon"`
	Color         string  `json:"color"`
	IsCompleted   bool    `json:"is_completed"`
	CreatedAt     string  `json:"created_at"`
}

// CreateGoalInput holds validated input for creating a savings goal.
type CreateGoalInput struct {
	Name         string  `json:"name"`
	TargetAmount float64 `json:"target_amount"`
	TargetDate   *string `json:"target_date"`
	Icon         string  `json:"icon"`
	Color        string  `json:"color"`
}

// UpdateGoalInput holds validated input for updating a savings goal.
type UpdateGoalInput struct {
	Name         string  `json:"name"`
	TargetAmount float64 `json:"target_amount"`
	TargetDate   *string `json:"target_date"`
	Icon         string  `json:"icon"`
	Color        string  `json:"color"`
}

// ContributeInput holds the amount to add to a savings goal.
type ContributeInput struct {
	Amount float64 `json:"amount"`
}
