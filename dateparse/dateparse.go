package dateparse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

var (
	naturalParserOnce sync.Once
	naturalParser     *when.Parser

	durationTokenRE = regexp.MustCompile(`(?i)(\d+)\s*([wdhms])`)
)

// ParseDateExpr parses absolute and natural language date expressions.
func ParseDateExpr(now time.Time, expr string) (time.Time, error) {
	text := strings.TrimSpace(expr)
	if text == "" {
		return time.Time{}, fmt.Errorf("date expression is empty")
	}

	if t, ok := parseAbsoluteDate(now.Location(), text); ok {
		return t, nil
	}

	r, err := getNaturalParser().Parse(text, now)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date expression %q: %w", expr, err)
	}
	if r != nil {
		return r.Time, nil
	}

	if t, ok := parseISOFallback(now.Location(), text); ok {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unsupported date expression %q", expr)
}

// ParseDurationExpr parses duration strings like 1d, 2h30m, and 1w.
func ParseDurationExpr(expr string) (time.Duration, error) {
	text := strings.TrimSpace(expr)
	if text == "" {
		return 0, fmt.Errorf("duration expression is empty")
	}

	matches := durationTokenRE.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration expression %q", expr)
	}

	var total time.Duration
	last := 0
	for _, m := range matches {
		start, end := m[0], m[1]
		if strings.TrimSpace(text[last:start]) != "" {
			return 0, fmt.Errorf("invalid duration expression %q", expr)
		}

		valueText := text[m[2]:m[3]]
		unitText := strings.ToLower(text[m[4]:m[5]])

		value, err := strconv.ParseInt(valueText, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration value %q: %w", valueText, err)
		}

		unitDuration, ok := durationUnit(unitText)
		if !ok {
			return 0, fmt.Errorf("unsupported duration unit %q", unitText)
		}

		total += time.Duration(value) * unitDuration
		last = end
	}

	if strings.TrimSpace(text[last:]) != "" {
		return 0, fmt.Errorf("invalid duration expression %q", expr)
	}

	return total, nil
}

func parseAbsoluteDate(loc *time.Location, expr string) (time.Time, bool) {
	layouts := []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, expr, loc); err == nil {
			return t, true
		}
	}

	if t, err := time.Parse(time.RFC3339, expr); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339Nano, expr); err == nil {
		return t, true
	}

	return time.Time{}, false
}

func parseISOFallback(loc *time.Location, expr string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339Nano, expr); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02", expr, loc); err == nil {
		return t, true
	}

	return time.Time{}, false
}

func getNaturalParser() *when.Parser {
	naturalParserOnce.Do(func() {
		parser := when.New(nil)
		parser.Add(en.All...)
		parser.Add(common.All...)
		naturalParser = parser
	})

	return naturalParser
}

func durationUnit(unit string) (time.Duration, bool) {
	switch unit {
	case "w":
		return 7 * 24 * time.Hour, true
	case "d":
		return 24 * time.Hour, true
	case "h":
		return time.Hour, true
	case "m":
		return time.Minute, true
	case "s":
		return time.Second, true
	default:
		return 0, false
	}
}
