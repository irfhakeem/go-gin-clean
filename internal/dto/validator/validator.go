package validator

import (
	stdErrors "errors"
	"fmt"
	"reflect"
	"strings"

	pkgerror "go-gin-clean/pkg/error"
	"go-gin-clean/pkg/response/message"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidations() error {
	engine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("failed to initialize validator engine")
	}

	engine.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "" && name != "-" {
			return name
		}
		if name := strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]; name != "" && name != "-" {
			return name
		}
		return fld.Name
	})

	if err := engine.RegisterValidation("gender", validateGender); err != nil {
		return err
	}

	if err := engine.RegisterValidation("password", validatePassword); err != nil {
		return err
	}

	return nil
}

func BuildValidationErrors(err error) (map[string]string, bool) {
	var ve validator.ValidationErrors
	if !stdErrors.As(err, &ve) {
		return nil, false
	}

	errs := make(map[string]string, len(ve))
	for _, fe := range ve {
		errs[fe.Field()] = buildFieldErrorMessage(fe)
	}

	return errs, true
}

func buildFieldErrorMessage(fe validator.FieldError) string {
	fieldName := toFriendlyName(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldName)
	case "email":
		return "must be a valid email address"
	case "min":
		switch fe.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
			return fmt.Sprintf("%s must be at least %s characters", fieldName, fe.Param())
		default:
			return fmt.Sprintf("%s must be at least %s", fieldName, fe.Param())
		}
	case "max":
		switch fe.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
			return fmt.Sprintf("%s must not exceed %s characters", fieldName, fe.Param())
		default:
			return fmt.Sprintf("%s must not exceed %s", fieldName, fe.Param())
		}
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", fieldName, fe.Param())
	case "oneof":
		opts := strings.ReplaceAll(fe.Param(), " ", ", ")
		return fmt.Sprintf("%s must be one of: %s", fieldName, opts)
	case "gte":
		return fmt.Sprintf("%s must be %s or greater", fieldName, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be %s or less", fieldName, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", fieldName, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", fieldName, fe.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fieldName)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fieldName)
	case "numeric":
		return fmt.Sprintf("%s must contain only numbers", fieldName)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", fieldName)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", fieldName)
	case "password":
		return message.Get(message.EN, pkgerror.ErrPasswordNotMeetsCriteria)
	case "gender":
		return message.Get(message.EN, pkgerror.ErrInvalidGender)
	default:
		return fmt.Sprintf("%s is invalid", fieldName)
	}
}

func toFriendlyName(field string) string {
	return strings.ReplaceAll(field, "_", " ")
}
