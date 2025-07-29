package enums

type Gender string

const (
	Male    Gender = "Male"
	Female  Gender = "Female"
	Unknown Gender = "Unknown"
)

// IsValid checks if the gender value is valid
func (g Gender) IsValid() bool {
	switch g {
	case Male, Female, Unknown:
		return true
	default:
		return false
	}
}

// String returns the string representation of gender
func (g Gender) String() string {
	return string(g)
}
