package validator

import (
	"go-gin-clean/internal/entity"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

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
