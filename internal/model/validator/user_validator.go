package validator

import (
	"go-gin-clean/internal/entity"
	"go-gin-clean/pkg/errors"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type UserValidator struct {
	validator *validator.Validate
}

func NewUserValidator() *UserValidator {
	v := validator.New()
	_ = v.RegisterValidation("gender", validateGender)
	_ = v.RegisterValidation("password", validatePassword)
	return &UserValidator{
		validator: v,
	}
}

func (cv *UserValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		for _, fieldError := range validationErrors {
			switch fieldError.Tag() {
			case "required":
				return errors.ErrInvalidInput
			case "email":
				return errors.ErrInvalidEmail
			case "min":
				if fieldError.Field() == "Password" {
					return errors.ErrPasswordNotMeetsCriteria
				}
				return errors.ErrInvalidInput
			case "password":
				return errors.ErrPasswordNotMeetsCriteria
			case "gender":
				return errors.ErrInvalidGender
			default:
				return errors.ErrInvalidInput
			}
		}
	}

	return nil
}

func validateGender(fl validator.FieldLevel) bool {
	value := fl.Field()

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return true
		}
		value = value.Elem()
	}

	if value.Type() != reflect.TypeOf(entity.Gender("")) {
		return false
	}

	gender := value.Interface().(entity.Gender)
	return gender.IsValid()
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[!@#~$%^&*()_+|<>{}[\]\/?]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSymbol
}
