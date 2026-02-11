// Package validate provides validation utilities for DTOs.
package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		if name != "" {
			return name
		}
		return fld.Name
	})
}

// ValidationError represents a validation failure with readable messages.
type ValidationError struct {
	Errors []FieldError
}

// FieldError represents a single field validation error.
type FieldError struct {
	Field   string
	Tag     string
	Value   interface{}
	Message string
}

// Error implements the error interface.
func (v *ValidationError) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}
	var msgs []string
	for _, e := range v.Errors {
		msgs = append(msgs, e.Message)
	}
	return strings.Join(msgs, "; ")
}

// Validate validates a struct and returns a ValidationError if validation fails.
func Validate(v interface{}) error {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	var fieldErrors []FieldError
	for _, fe := range validationErrors {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   fe.Field(),
			Tag:     fe.Tag(),
			Value:   fe.Value(),
			Message: formatErrorMessage(fe),
		})
	}

	return &ValidationError{Errors: fieldErrors}
}

func formatErrorMessage(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, fe.Param())
	case "required_with":
		return fmt.Sprintf("%s is required when %s is set", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s failed on '%s' validation", field, fe.Tag())
	}
}
