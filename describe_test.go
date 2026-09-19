package cronscribe

import "testing"

func TestDescribe(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"* * * * *", "every minute"},
		{"30 * * * *", "at minute 30 past every hour"},
		{"0 9 * * *", "at 09:00"},
		{"*/15 * * * *", "every 15 minutes"},
		{"* 9,17 * * *", "every minute during hours 9 and 17"},
		{"1,2,3 * * * *", "at minutes 1, 2 and 3 past every hour"},
		{"1,2 9,17 * * *", "at minute(s) 1 and 2 of hour(s) 9 and 17"},
		{"0 9 * * MON-FRI", "at 09:00, on Monday, Tuesday, Wednesday, Thursday and Friday"},
		{"0 0 1 1 *", "at 00:00, on day 1 of the month, in January"},
		{"0 0 1 * MON", "at 00:00, on day 1 of the month or on Monday"},
	}

	for _, c := range cases {
		s, err := Parse(c.expr)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", c.expr, err)
		}
		if got := s.Describe(); got != c.want {
			t.Errorf("Describe(%q) = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestEvenStep(t *testing.T) {
	cases := []struct {
		values   []int
		min, max int
		want     int
	}{
		{[]int{0, 15, 30, 45}, 0, 59, 15},
		{[]int{0, 10, 20, 30, 40, 50}, 0, 59, 10},
		{[]int{1, 2, 3}, 0, 59, 0}, // doesn't start at min
		{[]int{0, 1, 3}, 0, 59, 0}, // not constant step
		{[]int{0, 20}, 0, 59, 0},   // next value (40) would still fit, so this isn't a full */step run
		{[]int{5}, 0, 59, 0},       // single value
	}

	for _, c := range cases {
		if got := evenStep(c.values, c.min, c.max); got != c.want {
			t.Errorf("evenStep(%v, %d, %d) = %d, want %d", c.values, c.min, c.max, got, c.want)
		}
	}
}
