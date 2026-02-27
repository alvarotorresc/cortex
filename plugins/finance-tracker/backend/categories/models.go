package categories

// Category represents a transaction category with type filtering support.
type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	IsDefault bool   `json:"is_default"`
	SortOrder int    `json:"sort_order"`
}

// CreateCategoryInput holds the validated input for creating a category.
type CreateCategoryInput struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// UpdateCategoryInput holds the validated input for updating a category.
type UpdateCategoryInput struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// ReorderItem represents a single item in a reorder request.
type ReorderItem struct {
	ID        int64 `json:"id"`
	SortOrder int   `json:"sort_order"`
}

// validCategoryTypes defines the allowed category type values.
var validCategoryTypes = map[string]bool{
	"income":  true,
	"expense": true,
	"both":    true,
}

// IsValidCategoryType checks whether a given type string is allowed.
func IsValidCategoryType(categoryType string) bool {
	return validCategoryTypes[categoryType]
}
