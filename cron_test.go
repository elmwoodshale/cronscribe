package cronscribe

import "testing"

func TestParseWildcard(t *testing.T) {
	s, err := Parse("* * * * *")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Minute) != 60 || len(s.Hour) != 24 || len(s.DayOfMonth) != 31 ||
		len(s.Month) != 12 || len(s.DayOfWeek) != 7 {
		t.Fatalf("wildcard field did not expand to full range: %+v", s)
	}
}

func TestParseStepAndNamedRange(t *testing.T) {
	s, err := Parse("*/15 9-17 * * MON-FRI")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []int{0, 15, 30, 45}; !equalInts(s.Minute, want) {
		t.Errorf("minute = %v, want %v", s.Minute, want)
	}
	if want := []int{1, 2, 3, 4, 5}; !equalInts(s.DayOfWeek, want) {
		t.Errorf("day of week = %v, want %v", s.DayOfWeek, want)
	}
}

func TestParseSundayAlias(t *testing.T) {
	s, err := Parse("0 0 * * 7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []int{0}; !equalInts(s.DayOfWeek, want) {
		t.Errorf("day of week = %v, want %v", s.DayOfWeek, want)
	}
}

func TestParseRejectsBadInput(t *testing.T) {
	cases := []string{
		"60 * * * *",  // minute out of range
		"* 24 * * *",  // hour out of range
		"* * 0 * *",   // day of month starts at 1
		"* * * 13 *",  // month out of range
		"* * * * 8",   // day of week out of range
		"* * * *",     // too few fields
		"a b c d e",   // not numbers or names
		"5-1 * * * *", // descending range
	}
	for _, c := range cases {
		if _, err := Parse(c); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", c)
		}
	}
}

func TestFormatRoundTrip(t *testing.T) {
	cases := []string{
		"* * * * *",
		"0 0 1 1 0",
		"*/15 9-17 * * 1-5",
		"1,2,3,5,9-11 * * * *",
	}
	for _, c := range cases {
		s, err := Parse(c)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", c, err)
		}
		out := s.String()
		s2, err := Parse(out)
		if err != nil {
			t.Fatalf("Parse(%q), the round trip of %q, error: %v", out, c, err)
		}
		if s.String() != s2.String() {
			t.Errorf("round trip mismatch: %q -> %q -> %q", c, out, s2.String())
		}
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
