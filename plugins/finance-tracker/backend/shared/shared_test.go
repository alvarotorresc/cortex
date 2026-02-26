package shared

import (
	"encoding/json"
	"testing"

	_ "modernc.org/sqlite"
)

// --- AppError tests ---

func TestNewAppError(t *testing.T) {
	err := NewAppError("CUSTOM_ERROR", "something went wrong", 500)

	if err.Code != "CUSTOM_ERROR" {
		t.Errorf("expected code 'CUSTOM_ERROR', got '%s'", err.Code)
	}
	if err.Message != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got '%s'", err.Message)
	}
	if err.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", err.StatusCode)
	}
	if err.Error() != "something went wrong" {
		t.Errorf("expected Error() to return message, got '%s'", err.Error())
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("field is required")

	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got '%s'", err.Code)
	}
	if err.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", err.StatusCode)
	}
	if err.Message != "field is required" {
		t.Errorf("expected message 'field is required', got '%s'", err.Message)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("transaction", "42")

	if err.Code != "NOT_FOUND" {
		t.Errorf("expected code 'NOT_FOUND', got '%s'", err.Code)
	}
	if err.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
	if err.Message != "transaction 42 not found" {
		t.Errorf("expected message 'transaction 42 not found', got '%s'", err.Message)
	}
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("resource already exists")

	if err.Code != "CONFLICT" {
		t.Errorf("expected code 'CONFLICT', got '%s'", err.Code)
	}
	if err.StatusCode != 409 {
		t.Errorf("expected status 409, got %d", err.StatusCode)
	}
	if err.Message != "resource already exists" {
		t.Errorf("expected message 'resource already exists', got '%s'", err.Message)
	}
}

// --- JSONSuccess tests ---

func TestJSONSuccess(t *testing.T) {
	data := map[string]string{"name": "test"}

	resp, err := JSONSuccess(200, data)
	if err != nil {
		t.Fatalf("JSONSuccess returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.ContentType != "application/json" {
		t.Errorf("expected content type 'application/json', got '%s'", resp.ContentType)
	}

	var body struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if body.Data["name"] != "test" {
		t.Errorf("expected data.name to be 'test', got '%s'", body.Data["name"])
	}
}

func TestJSONSuccess_WithSlice(t *testing.T) {
	data := []string{"one", "two"}

	resp, err := JSONSuccess(201, data)
	if err != nil {
		t.Fatalf("JSONSuccess returned error: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var body struct {
		Data []string `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if len(body.Data) != 2 {
		t.Errorf("expected 2 items, got %d", len(body.Data))
	}
}

// --- JSONError tests ---

func TestJSONError(t *testing.T) {
	appErr := NewValidationError("invalid input")

	resp, err := JSONError(appErr)
	if err != nil {
		t.Fatalf("JSONError returned error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
	if resp.ContentType != "application/json" {
		t.Errorf("expected content type 'application/json', got '%s'", resp.ContentType)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to unmarshal error body: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected error code 'VALIDATION_ERROR', got '%s'", body.Error.Code)
	}
	if body.Error.Message != "invalid input" {
		t.Errorf("expected error message 'invalid input', got '%s'", body.Error.Message)
	}
}

func TestJSONError_NotFound(t *testing.T) {
	appErr := NewNotFoundError("account", "7")

	resp, err := JSONError(appErr)
	if err != nil {
		t.Fatalf("JSONError returned error: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to unmarshal error body: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code 'NOT_FOUND', got '%s'", body.Error.Code)
	}
	if body.Error.Message != "account 7 not found" {
		t.Errorf("expected message 'account 7 not found', got '%s'", body.Error.Message)
	}
}

// --- ExtractIDFromPath tests ---

func TestExtractIDFromPath_Valid(t *testing.T) {
	id, appErr := ExtractIDFromPath("/transactions/42")
	if appErr != nil {
		t.Fatalf("expected no error, got: %s", appErr.Message)
	}
	if id != 42 {
		t.Errorf("expected id 42, got %d", id)
	}
}

func TestExtractIDFromPath_NestedPath(t *testing.T) {
	id, appErr := ExtractIDFromPath("/accounts/7/transactions")
	if appErr != nil {
		t.Fatalf("expected no error, got: %s", appErr.Message)
	}
	if id != 7 {
		t.Errorf("expected id 7, got %d", id)
	}
}

func TestExtractIDFromPath_MissingID(t *testing.T) {
	_, appErr := ExtractIDFromPath("/transactions/")
	if appErr == nil {
		t.Fatal("expected error for missing ID, got nil")
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got '%s'", appErr.Code)
	}
	if appErr.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", appErr.StatusCode)
	}
}

func TestExtractIDFromPath_NonNumeric(t *testing.T) {
	_, appErr := ExtractIDFromPath("/transactions/abc")
	if appErr == nil {
		t.Fatal("expected error for non-numeric ID, got nil")
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got '%s'", appErr.Code)
	}
	if appErr.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", appErr.StatusCode)
	}
}

func TestExtractIDFromPath_SingleSegment(t *testing.T) {
	_, appErr := ExtractIDFromPath("/transactions")
	if appErr == nil {
		t.Fatal("expected error for path with no ID segment, got nil")
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code 'VALIDATION_ERROR', got '%s'", appErr.Code)
	}
}
