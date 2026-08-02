package validator

import (
	"reflect"
	"regexp"

	"go-gin-clean/internal/entity"

	"github.com/go-playground/validator/v10"
)

var (
	reUpper  = regexp.MustCompile(`[A-Z]`)
	reLower  = regexp.MustCompile(`[a-z]`)
	reNumber = regexp.MustCompile(`[0-9]`)
	reSymbol = regexp.MustCompile(`[!@#~$%^&*()\-_+|<>{}[\]\/?]`)
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

	return value.Interface().(entity.Gender).IsValid()
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) >= 8 &&
		reUpper.MatchString(password) &&
		reLower.MatchString(password) &&
		reNumber.MatchString(password) &&
		reSymbol.MatchString(password)
}
