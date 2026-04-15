package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat rule is empty")
	}

	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("invalid dstart format: %w", err)
	}

	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	parts := strings.Split(repeat, " ")
	ruleType := parts[0]

	switch ruleType {
	case "d":
		return nextDateDaily(now, date, parts)
	case "y":
		return nextDateYearly(now, date)
	case "w":
		return nextDateWeekly(now, date, parts)
	case "m":
		return nextDateMonthly(now, date, parts)
	default:
		return "", fmt.Errorf("unsupported repeat rule: %s", ruleType)
	}
}

func nextDateDaily(now time.Time, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", errors.New("daily rule requires format: d <number>")
	}

	interval, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid daily interval: %s", parts[1])
	}

	if interval < 1 || interval > 400 {
		return "", fmt.Errorf("daily interval must be between 1 and 400, got %d", interval)
	}

	for {
		date = date.AddDate(0, 0, interval)
		if date.After(now) {
			return date.Format(DateFormat), nil
		}
	}
}

func nextDateYearly(now time.Time, date time.Time) (string, error) {
	for {
		date = date.AddDate(1, 0, 0)
		if date.After(now) {
			return date.Format(DateFormat), nil
		}
	}
}

func nextDateWeekly(now time.Time, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", errors.New("weekly rule requires format: w <days>")
	}

	daysStr := strings.Split(parts[1], ",")
	var allowedDays [8]bool

	for _, d := range daysStr {
		day, err := strconv.Atoi(strings.TrimSpace(d))
		if err != nil {
			return "", fmt.Errorf("invalid day in weekly rule: %s", d)
		}
		if day < 1 || day > 7 {
			return "", fmt.Errorf("day must be between 1 and 7, got %d", day)
		}
		allowedDays[day] = true
	}

	for {
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7 // вс
		}

		if allowedDays[weekday] && date.After(now) {
			return date.Format(DateFormat), nil
		}
		date = date.AddDate(0, 0, 1)
	}
}

func nextDateMonthly(now time.Time, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 || len(parts) > 3 {
		return "", errors.New("monthly rule requires format: m <days> [months]")
	}

	daysStr := strings.Split(parts[1], ",")
	var allowedDays [32]bool
	checkLast := false
	checkLastMinusOne := false

	for _, d := range daysStr {
		dayStr := strings.TrimSpace(d)
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return "", fmt.Errorf("invalid day: %s", d)
		}
		if day == -1 {
			checkLast = true
		} else if day == -2 {
			checkLastMinusOne = true
		} else if day < 1 || day > 31 {
			return "", fmt.Errorf("day must be between 1 and 31, or -1, -2, got %d", day)
		} else {
			allowedDays[day] = true
		}
	}

	var allowedMonths [13]bool
	monthSpecified := false
	if len(parts) == 3 {
		monthsStr := strings.Split(parts[2], ",")
		for _, m := range monthsStr {
			month, err := strconv.Atoi(strings.TrimSpace(m))
			if err != nil {
				return "", fmt.Errorf("invalid month: %s", m)
			}
			if month < 1 || month > 12 {
				return "", fmt.Errorf("month must be between 1 and 12, got %d", month)
			}
			allowedMonths[month] = true
		}
		monthSpecified = true
	}

	for {
		if monthSpecified && !allowedMonths[int(date.Month())] {
			date = date.AddDate(0, 0, 1)
			continue
		}

		day := date.Day()
		lastDay := lastDayOfMonth(date)

		matched := false
		if checkLast && day == lastDay {
			matched = true
		} else if checkLastMinusOne && day == lastDay-1 {
			matched = true
		} else if allowedDays[day] {
			matched = true
		}

		if matched && date.After(now) {
			return date.Format(DateFormat), nil
		}

		date = date.AddDate(0, 0, 1)
	}
}

func lastDayOfMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeatStr := r.FormValue("repeat")

	if dateStr == "" {
		http.Error(w, `{"error":"date parameter is required"}`, http.StatusBadRequest)
		return
	}
	if repeatStr == "" {
		http.Error(w, `{"error":"repeat parameter is required"}`, http.StatusBadRequest)
		return
	}

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			http.Error(w, `{"error":"invalid now format, expected 20060102"}`, http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
