package tags

import (
	"testing"
)

func TestValidateCreateInput_MissingName(t *testing.T) {
	input := &CreateTagInput{Name: "", Color: "#FF0000"}
	if err := validateCreateInput(input); err == nil {
		t.Error("expected validation error for empty name")
	}
}

func TestValidateCreateInput_WhitespaceName(t *testing.T) {
	input := &CreateTagInput{Name: "   ", Color: "#FF0000"}
	if err := validateCreateInput(input); err == nil {
		t.Error("expected validation error for whitespace-only name")
	}
}

func TestValidateCreateInput_Valid(t *testing.T) {
	input := &CreateTagInput{Name: "groceries", Color: "#00FF00"}
	if err := validateCreateInput(input); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestValidateCreateInput_ValidNoColor(t *testing.T) {
	input := &CreateTagInput{Name: "travel"}
	if err := validateCreateInput(input); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestValidateUpdateInput_MissingName(t *testing.T) {
	input := &UpdateTagInput{Name: "", Color: "#FF0000"}
	if err := validateUpdateInput(input); err == nil {
		t.Error("expected validation error for empty name")
	}
}

func TestValidateUpdateInput_Valid(t *testing.T) {
	input := &UpdateTagInput{Name: "updated", Color: "#0000FF"}
	if err := validateUpdateInput(input); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}
