package validate

import (
	"testing"
)

func TestValidate_RequiredField(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	err := Validate(&TestStruct{Name: ""})
	if err == nil {
		t.Fatal("expected validation error for empty required field")
	}

	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if len(valErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(valErr.Errors))
	}

	if valErr.Errors[0].Tag != "required" {
		t.Errorf("expected tag 'required', got %q", valErr.Errors[0].Tag)
	}
}

func TestValidate_MinField(t *testing.T) {
	type TestStruct struct {
		Count int `validate:"min=0"`
	}

	err := Validate(&TestStruct{Count: -1})
	if err == nil {
		t.Fatal("expected validation error for value below min")
	}

	err = Validate(&TestStruct{Count: 0})
	if err != nil {
		t.Fatalf("expected no error for value at min, got %v", err)
	}
}

func TestValidate_OneOfField(t *testing.T) {
	type TestStruct struct {
		Severity string `validate:"oneof=critical high warning info"`
	}

	err := Validate(&TestStruct{Severity: "high"})
	if err != nil {
		t.Fatalf("expected no error for valid oneof value, got %v", err)
	}

	err = Validate(&TestStruct{Severity: "invalid"})
	if err == nil {
		t.Fatal("expected validation error for invalid oneof value")
	}
}

func TestValidate_RequiredWithField(t *testing.T) {
	type TestStruct struct {
		HeadRef string `json:"head_ref"`
		BaseRef string `json:"base_ref" validate:"required_with=HeadRef"`
	}

	err := Validate(&TestStruct{HeadRef: "main", BaseRef: ""})
	if err == nil {
		t.Fatal("expected validation error when HeadRef is set but BaseRef is empty")
	}

	err = Validate(&TestStruct{HeadRef: "", BaseRef: ""})
	if err != nil {
		t.Fatalf("expected no error when both are empty, got %v", err)
	}

	err = Validate(&TestStruct{HeadRef: "main", BaseRef: "develop"})
	if err != nil {
		t.Fatalf("expected no error when both are set, got %v", err)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Count int    `validate:"min=0"`
	}

	err := Validate(&TestStruct{Name: "", Count: -1})
	if err == nil {
		t.Fatal("expected validation errors")
	}

	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}

	if len(valErr.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(valErr.Errors))
	}
}

func TestValidate_ValidStruct(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Count int    `validate:"min=0"`
	}

	err := Validate(&TestStruct{Name: "test", Count: 5})
	if err != nil {
		t.Fatalf("expected no error for valid struct, got %v", err)
	}
}

func TestValidationError_Error(t *testing.T) {
	valErr := &ValidationError{
		Errors: []FieldError{
			{Field: "name", Tag: "required", Message: "name is required"},
			{Field: "count", Tag: "min", Message: "count must be at least 0"},
		},
	}

	errMsg := valErr.Error()
	if errMsg != "name is required; count must be at least 0" {
		t.Errorf("unexpected error message: %s", errMsg)
	}
}

func TestValidationError_EmptyErrors(t *testing.T) {
	valErr := &ValidationError{}
	if valErr.Error() != "validation failed" {
		t.Errorf("expected 'validation failed', got %q", valErr.Error())
	}
}

func TestValidate_JSONTagUsed(t *testing.T) {
	type TestStruct struct {
		MyField string `json:"my_field" validate:"required"`
	}

	err := Validate(&TestStruct{MyField: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}

	valErr := err.(*ValidationError)
	if valErr.Errors[0].Field != "my_field" {
		t.Errorf("expected field name 'my_field', got %q", valErr.Errors[0].Field)
	}
}
