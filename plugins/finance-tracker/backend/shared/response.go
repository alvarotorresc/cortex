package shared

import (
	"encoding/json"
	"fmt"

	"github.com/alvarotorresc/cortex/pkg/sdk"
)

// JSONSuccess wraps data in {"data": ...} format per PATTERNS.md and returns
// an APIResponse with the given HTTP status code.
func JSONSuccess(status int, data interface{}) (*sdk.APIResponse, error) {
	body, err := json.Marshal(map[string]interface{}{"data": data})
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}
	return &sdk.APIResponse{
		StatusCode:  status,
		Body:        body,
		ContentType: "application/json",
	}, nil
}

// JSONError converts an AppError into a standardized error response with
// {"error": {"code": ..., "message": ...}} format per PATTERNS.md.
func JSONError(appErr *AppError) (*sdk.APIResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	})
	return &sdk.APIResponse{
		StatusCode:  appErr.StatusCode,
		Body:        body,
		ContentType: "application/json",
	}, nil
}
