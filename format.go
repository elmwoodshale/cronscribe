package cronscribe

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// String renders the schedule back into a canonical five-field cron
// expression, collapsing runs of consecutive values into ranges. It only
// reads its receiver, so it works the same whether the Schedule came
// from Parse or was built by hand.
func (s Schedule) String() string {
	fields := []string{
		formatField(s.Minute, 0, 59),
		formatField(s.Hour, 0, 23),
		formatField(s.DayOfMonth, 1, 31),
		formatField(s.Month, 1, 12),
		formatField(s.DayOfWeek, 0, 6),
	}
	return strings.Join(fields, " ")
}

func formatField(values []int, min, max int) string {
	if len(values) == 0 {
		return "*"
	}

	uniq := dedupeSorted(values)

	if isFullRange(uniq, min, max) {
		return "*"
	}

	var parts []string
	for i := 0; i < len(uniq); {
		start := uniq[i]
		end := start
		j := i + 1
		for j < len(uniq) && uniq[j] == end+1 {
			end = uniq[j]
			j++
		}
		switch {
		case start == end:
			parts = append(parts, strconv.Itoa(start))
		case end == start+1:
			parts = append(parts, strconv.Itoa(start), strconv.Itoa(end))
		default:
			parts = append(parts, fmt.Sprintf("%d-%d", start, end))
		}
		i = j
	}

	return strings.Join(parts, ",")
}

func dedupeSorted(values []int) []int {
	cp := append([]int(nil), values...)
	sort.Ints(cp)
	out := cp[:0]
	for i, v := range cp {
		if i == 0 || v != out[len(out)-1] {
			out = append(out, v)
		}
	}
	return out
}

func isFullRange(sorted []int, min, max int) bool {
	if len(sorted) != max-min+1 {
		return false
	}
	for i, v := range sorted {
		if v != min+i {
			return false
		}
	}
	return true
}
