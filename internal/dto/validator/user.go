package validator

import (
	"go-gin-clean/internal/domain/vo"
	"reflect"
	"regexp"

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

	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return true
		}
		value = value.Elem()
	}

	return vo.IsValidGender(value.String())
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) >= 8 &&
		reUpper.MatchString(password) &&
		reLower.MatchString(password) &&
		reNumber.MatchString(password) &&
		reSymbol.MatchString(password)
}
