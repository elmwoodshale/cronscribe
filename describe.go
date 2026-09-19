package cronscribe

import (
	"fmt"
	"strconv"
	"strings"
)

var monthLabels = [...]string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}

var weekdayLabels = [...]string{
	"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
}

// Describe renders the schedule as an English sentence describing when it
// runs. It recognizes common shapes - a fixed time of day, an hourly
// interval, a fixed number of minutes - and falls back to naming the raw
// field values for anything that doesn't fit one of those shapes.
func (s Schedule) Describe() string {
	minuteFull := isFullRange(dedupeSorted(s.Minute), 0, 59)
	hourFull := isFullRange(dedupeSorted(s.Hour), 0, 23)
	domFull := isFullRange(dedupeSorted(s.DayOfMonth), 1, 31)
	monthFull := isFullRange(dedupeSorted(s.Month), 1, 12)
	dowFull := isFullRange(dedupeSorted(s.DayOfWeek), 0, 6)

	parts := []string{describeTime(s, minuteFull, hourFull)}

	switch {
	case !domFull && !dowFull:
		// Standard cron treats a restriction on both fields as "day of
		// month OR day of week", not an intersection of the two.
		parts = append(parts, fmt.Sprintf("%s or %s", describeDayOfMonth(s), describeDayOfWeek(s)))
	case !domFull:
		parts = append(parts, describeDayOfMonth(s))
	case !dowFull:
		parts = append(parts, describeDayOfWeek(s))
	}

	if !monthFull {
		parts = append(parts, fmt.Sprintf("in %s", joinLabels(s.Month, monthLabels[:], 1)))
	}

	return strings.Join(parts, ", ")
}

func describeTime(s Schedule, minuteFull, hourFull bool) string {
	switch {
	case minuteFull && hourFull:
		return "every minute"
	case minuteFull:
		return fmt.Sprintf("every minute during %s %s", pluralize(len(s.Hour), "hour", "hours"), joinInts(s.Hour))
	case len(s.Minute) == 1 && len(s.Hour) == 1:
		return fmt.Sprintf("at %02d:%02d", s.Hour[0], s.Minute[0])
	case len(s.Minute) == 1 && hourFull:
		return fmt.Sprintf("at minute %d past every hour", s.Minute[0])
	}

	if step := evenStep(s.Minute, 0, 59); step > 0 && hourFull {
		return fmt.Sprintf("every %d minutes", step)
	}
	if hourFull {
		return fmt.Sprintf("at minutes %s past every hour", joinInts(s.Minute))
	}
	return fmt.Sprintf("at minute(s) %s of hour(s) %s", joinInts(s.Minute), joinInts(s.Hour))
}

func describeDayOfMonth(s Schedule) string {
	return fmt.Sprintf("on %s %s of the month", pluralize(len(s.DayOfMonth), "day", "days"), joinInts(s.DayOfMonth))
}

func describeDayOfWeek(s Schedule) string {
	return fmt.Sprintf("on %s", joinLabels(s.DayOfWeek, weekdayLabels[:], 0))
}

// evenStep reports the step size if values is exactly the arithmetic
// sequence min, min+step, min+2*step, ... produced by parsing "*/step",
// stopping before it would exceed max. It returns 0 for anything else.
func evenStep(values []int, min, max int) int {
	if len(values) < 2 || values[0] != min {
		return 0
	}
	step := values[1] - values[0]
	if step <= 0 {
		return 0
	}
	for i := 2; i < len(values); i++ {
		if values[i]-values[i-1] != step {
			return 0
		}
	}
	if values[len(values)-1]+step <= max {
		return 0
	}
	return step
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func joinInts(values []int) string {
	items := make([]string, len(values))
	for i, v := range values {
		items[i] = strconv.Itoa(v)
	}
	return joinWithAnd(items)
}

// joinLabels renders values as names, e.g. weekday or month labels. offset
// is subtracted from each value before indexing into labels, since months
// are 1-indexed and weekdays are 0-indexed.
func joinLabels(values []int, labels []string, offset int) string {
	items := make([]string, len(values))
	for i, v := range values {
		items[i] = labels[v-offset]
	}
	return joinWithAnd(items)
}

func joinWithAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
	}
}
