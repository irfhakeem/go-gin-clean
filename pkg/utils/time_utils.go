package utils

import "time"

func FormatTimeToString(t time.Time, format string, zone string) string {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		loc = time.UTC
	}
	return t.In(loc).Format(format)
}

func ParseStringToTime(tStr string, format string, zone *string) (time.Time, error) {
	loc := time.UTC
	if zone != nil {
		location, err := time.LoadLocation(*zone)
		if err == nil {
			loc = location
		}
	}
	return time.ParseInLocation(format, tStr, loc)
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
