package parser

import (
	"fmt"
	"strings"
	"time"
)

const MealTimezone = "Europe/Kyiv"

func MealLocation() *time.Location {
	loc, err := time.LoadLocation(MealTimezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func ResolveEatenAt(raw string, now time.Time) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}

	loc := MealLocation()
	nowLocal := now.In(loc)

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	var parsed time.Time
	var err error
	for _, layout := range layouts {
		parsed, err = time.ParseInLocation(layout, raw, loc)
		if err == nil {
			if layout == "2006-01-02" {
				parsed = time.Date(
					parsed.Year(), parsed.Month(), parsed.Day(),
					nowLocal.Hour(), nowLocal.Minute(), nowLocal.Second(),
					nowLocal.Nanosecond(), loc,
				)
			}
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid eaten_at %q", raw)
}
