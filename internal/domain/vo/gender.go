package vo

import "errors"

var ErrInvalidGender = errors.New("invalid gender")

type Gender string

const (
	GenderMale     Gender = "male"
	GenderFemale   Gender = "female"
	GenderNotToSay Gender = "not_to_say"
)

func (g Gender) String() string {
	return string(g)
}

func ParseGender(gender string) (Gender, error) {
	switch Gender(gender) {
	case GenderMale:
		return GenderMale, nil
	case GenderFemale:
		return GenderFemale, nil
	case GenderNotToSay:
		return GenderNotToSay, nil
	default:
		return "", ErrInvalidGender
	}
}

func IsValidGender(gender string) bool {
	switch Gender(gender) {
	case GenderMale, GenderFemale, GenderNotToSay:
		return true
	default:
		return false
	}
}
