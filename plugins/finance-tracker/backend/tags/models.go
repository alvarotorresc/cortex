package tags

// Tag represents a label that can be attached to transactions.
type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CreateTagInput holds the validated input for creating a tag.
type CreateTagInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// UpdateTagInput holds the validated input for updating a tag.
type UpdateTagInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
