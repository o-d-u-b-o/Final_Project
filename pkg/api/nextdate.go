package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	dateFormat = "20060102"
	maxDays    = 400
)

var (
	ErrEmptyRepeat     = errors.New("empty repeat rule")
	ErrInvalidFormat   = errors.New("invalid repeat format")
	ErrInvalidDay      = errors.New("invalid day")
	ErrInvalidMonth    = errors.New("invalid month")
	ErrInvalidWeekday  = errors.New("invalid weekday")
	ErrMaxDaysExceeded = errors.New("max days exceeded")
	ErrUnsupportedRule = errors.New("unsupported repeat rule")
)

func NextDate(now time.Time, dateStr, repeat string) (string, error) {
	if repeat == "" {
		return "", ErrEmptyRepeat
	}

	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date: %w", err)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", ErrInvalidFormat
	}

	switch parts[0] {
	case "d":
		return handleDailyRule(now, date, parts)
	case "y":
		return handleYearlyRule(now, date), nil
	case "w":
		return handleWeeklyRule(now, date, parts)
	case "m":
		return handleMonthlyRule(now, date, parts)
	default:
		return "", ErrUnsupportedRule
	}
}

func handleDailyRule(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", ErrInvalidFormat
	}

	days, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", ErrInvalidFormat
	}

	if days <= 0 || days > maxDays {
		return "", ErrMaxDaysExceeded
	}

	for {
		date = date.AddDate(0, 0, days)
		if afterNow(date, now) {
			break
		}
	}

	return date.Format(dateFormat), nil
}

func handleYearlyRule(now, date time.Time) string {
	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			break
		}
	}
	return date.Format(dateFormat)
}

func handleWeeklyRule(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", ErrInvalidFormat
	}

	weekdays := make(map[int]bool)
	for _, dayStr := range strings.Split(parts[1], ",") {
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 7 {
			return "", ErrInvalidWeekday
		}
		weekdays[day] = true
	}

	for {
		date = date.AddDate(0, 0, 1)
		if afterNow(date, now) {
			weekday := int(date.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			if weekdays[weekday] {
				break
			}
		}
	}

	return date.Format(dateFormat), nil
}

func handleMonthlyRule(now, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 || len(parts) > 3 {
		return "", ErrInvalidFormat
	}

	days, err := parseDays(parts[1])
	if err != nil {
		return "", err
	}

	var months map[int]bool
	if len(parts) == 3 {
		months, err = parseMonths(parts[2])
		if err != nil {
			return "", err
		}
	}

	for {
		date = date.AddDate(0, 0, 1)
		if afterNow(date, now) {
			day := date.Day()
			month := int(date.Month())

			if days[-1] && isLastDayOfMonth(date) {
				return date.Format(dateFormat), nil
			}
			if days[-2] && isPenultimateDayOfMonth(date) {
				return date.Format(dateFormat), nil
			}

			if (months == nil || months[month]) && days[day] {
				return date.Format(dateFormat), nil
			}
		}
	}
}

func afterNow(date, now time.Time) bool {
	return date.After(now)
}

func isLastDayOfMonth(date time.Time) bool {
	return date.AddDate(0, 0, 1).Month() != date.Month()
}

func isPenultimateDayOfMonth(date time.Time) bool {
	return date.AddDate(0, 0, 2).Month() != date.Month()
}

func parseDays(s string) (map[int]bool, error) {
	days := make(map[int]bool)
	for _, dayStr := range strings.Split(s, ",") {
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			if dayStr == "-1" {
				days[-1] = true
				continue
			}
			if dayStr == "-2" {
				days[-2] = true
				continue
			}
			return nil, ErrInvalidDay
		}
		if day < 1 || day > 31 {
			return nil, ErrInvalidDay
		}
		days[day] = true
	}
	return days, nil
}

func parseMonths(s string) (map[int]bool, error) {
	months := make(map[int]bool)
	for _, monthStr := range strings.Split(s, ",") {
		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			return nil, ErrInvalidMonth
		}
		months[month] = true
	}
	return months, nil
}
