package validator

import (
	stdErrors "errors"
	"fmt"
	"go-gin-clean/pkg/errors"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidations() error {
	engine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("failed to initialize validator engine")
	}

	// Register custom validations func here:
	if err := engine.RegisterValidation("gender", validateGender); err != nil {
		return err
	}

	if err := engine.RegisterValidation("password", validatePassword); err != nil {
		return err
	}

	return nil
}

func BuildValidationMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !stdErrors.As(err, &validationErrors) {
		return err.Error()
	}

	messages := make([]string, 0, len(validationErrors))
	for _, fieldError := range validationErrors {
		messages = append(messages, formatFieldErrorMessage(fieldError))
	}

	return strings.Join(messages, "; ")
}

func formatFieldErrorMessage(fieldError validator.FieldError) string {
	fieldName := toFriendlyFieldName(fieldError.Field())

	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fieldName, fieldError.Param())
	case "email":
		return errors.ErrInvalidEmail.Error()
	case "password":
		return errors.ErrPasswordNotMeetsCriteria.Error()
	case "gender":
		return errors.ErrInvalidGender.Error()
	default:
		return fmt.Sprintf("%s is invalid", fieldName)
	}
}

func toFriendlyFieldName(field string) string {
	if field == "" {
		return "field"
	}

	var builder strings.Builder
	for i, r := range field {
		if i > 0 && unicode.IsUpper(r) {
			builder.WriteRune(' ')
		}
		builder.WriteRune(unicode.ToLower(r))
	}

	return builder.String()
}
