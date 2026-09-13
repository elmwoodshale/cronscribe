// Package cronscribe parses and formats standard five-field cron
// expressions (minute hour day-of-month month day-of-week).
package cronscribe

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Schedule is the parsed, validated form of a cron expression. Each slice
// holds the sorted, deduplicated set of values that field matches; an
// empty slice never occurs for a Schedule produced by Parse.
type Schedule struct {
	Minute     []int
	Hour       []int
	DayOfMonth []int
	Month      []int
	DayOfWeek  []int
}

type fieldSpec struct {
	name  string
	min   int
	max   int
	names map[string]int
	alias map[int]int
}

var monthNames = map[string]int{
	"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
	"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
}

var dowNames = map[string]int{
	"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
}

// Parse validates a standard five-field cron expression and returns the
// resolved schedule. It accepts *, lists (1,2,3), ranges (1-5), steps
// (*/15, 1-10/2) and, for month and day-of-week, three-letter names
// (JAN, MON). It does not accept the "@daily" style shorthands, a
// seconds field, or descending ranges that wrap around (22-2).
func Parse(expr string) (Schedule, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return Schedule{}, fmt.Errorf("cron: expected 5 fields, got %d", len(fields))
	}

	specs := [5]fieldSpec{
		{name: "minute", min: 0, max: 59},
		{name: "hour", min: 0, max: 23},
		{name: "day of month", min: 1, max: 31},
		{name: "month", min: 1, max: 12, names: monthNames},
		{name: "day of week", min: 0, max: 7, names: dowNames, alias: map[int]int{7: 0}},
	}

	var s Schedule
	targets := [5]*[]int{&s.Minute, &s.Hour, &s.DayOfMonth, &s.Month, &s.DayOfWeek}

	for i, spec := range specs {
		values, err := parseField(fields[i], spec)
		if err != nil {
			return Schedule{}, fmt.Errorf("cron: %s field %q: %w", spec.name, fields[i], err)
		}
		*targets[i] = values
	}

	return s, nil
}

func parseField(raw string, spec fieldSpec) ([]int, error) {
	seen := make(map[int]bool)
	var values []int

	for _, part := range strings.Split(raw, ",") {
		if part == "" {
			return nil, fmt.Errorf("empty item in list")
		}

		base, step, err := splitStep(part)
		if err != nil {
			return nil, err
		}

		lo, hi, err := resolveRange(base, spec)
		if err != nil {
			return nil, err
		}

		for v := lo; v <= hi; v += step {
			if v < spec.min || v > spec.max {
				return nil, fmt.Errorf("value %d out of range [%d,%d]", v, spec.min, spec.max)
			}
			if a, ok := spec.alias[v]; ok {
				v = a
			}
			if !seen[v] {
				seen[v] = true
				values = append(values, v)
			}
		}
	}

	sort.Ints(values)
	return values, nil
}

func splitStep(part string) (base string, step int, err error) {
	pieces := strings.Split(part, "/")
	switch len(pieces) {
	case 1:
		return pieces[0], 1, nil
	case 2:
		step, err = strconv.Atoi(pieces[1])
		if err != nil || step <= 0 {
			return "", 0, fmt.Errorf("invalid step %q", pieces[1])
		}
		return pieces[0], step, nil
	default:
		return "", 0, fmt.Errorf("too many '/' in %q", part)
	}
}

func resolveRange(base string, spec fieldSpec) (lo, hi int, err error) {
	if base == "*" {
		return spec.min, spec.max, nil
	}

	bounds := strings.SplitN(base, "-", 2)
	if len(bounds) == 1 {
		v, err := resolveToken(bounds[0], spec)
		if err != nil {
			return 0, 0, err
		}
		return v, v, nil
	}

	lo, err = resolveToken(bounds[0], spec)
	if err != nil {
		return 0, 0, err
	}
	hi, err = resolveToken(bounds[1], spec)
	if err != nil {
		return 0, 0, err
	}
	if lo > hi {
		return 0, 0, fmt.Errorf("range %q descends, start must be <= end", base)
	}
	return lo, hi, nil
}

func resolveToken(tok string, spec fieldSpec) (int, error) {
	if spec.names != nil {
		if v, ok := spec.names[strings.ToUpper(tok)]; ok {
			return v, nil
		}
	}
	v, err := strconv.Atoi(tok)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q", tok)
	}
	return v, nil
}
