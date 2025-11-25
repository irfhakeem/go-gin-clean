package utils

import "time"

func FormatTimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

func ParseStringToTime(tStr string) (time.Time, error) {
	layouts := []string{
		time.RFC3339, // ex: 2025-02-15T10:30:00Z
		"2006-01-02", // ex: 2025-02-15
	}

	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, tStr)
		if err == nil {
			break
		}
	}

	return t, err
}

func SetDefaultTime(t time.Time) time.Time {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		return t.Add(0)
	}
	return t
}

func GetCurrentTime(timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc)
}
