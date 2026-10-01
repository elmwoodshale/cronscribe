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

func TestParseShorthands(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"@yearly", "0 0 1 1 *"},
		{"@annually", "0 0 1 1 *"},
		{"@monthly", "0 0 1 * *"},
		{"@weekly", "0 0 * * 0"},
		{"@daily", "0 0 * * *"},
		{"@midnight", "0 0 * * *"},
		{"@hourly", "0 * * * *"},
	}
	for _, c := range cases {
		s, err := Parse(c.expr)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", c.expr, err)
		}
		want, err := Parse(c.want)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", c.want, err)
		}
		if s.String() != want.String() {
			t.Errorf("Parse(%q) = %q, want %q", c.expr, s.String(), want.String())
		}
	}
}

func TestParseRejectsBadInput(t *testing.T) {
	cases := []string{
		"60 * * * *",   // minute out of range
		"* 24 * * *",   // hour out of range
		"* * 0 * *",    // day of month starts at 1
		"* * * 13 *",   // month out of range
		"* * * * 8",    // day of week out of range
		"* * * *",      // too few fields
		"a b c d e",    // not numbers or names
		"5-1 * * * *",  // descending range
		"@fortnightly", // not a recognized shorthand
	}
	for _, c := range cases {
		if _, err := Parse(c); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", c)
		}
	}
}

func TestParseSecondsField(t *testing.T) {
	s, err := Parse("*/20 30 9 * * MON")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !equalInts(s.Second, []int{0, 20, 40}) {
		t.Errorf("Second = %v, want [0 20 40]", s.Second)
	}
	if !equalInts(s.Minute, []int{30}) || !equalInts(s.Hour, []int{9}) || !equalInts(s.DayOfWeek, []int{1}) {
		t.Errorf("other fields shifted wrongly: %+v", s)
	}
	if got, want := s.String(), "0,20,40 30 9 * * 1"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestParseFiveFieldHasNoSeconds(t *testing.T) {
	s, err := Parse("0 9 * * *")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if s.Second != nil {
		t.Errorf("Second = %v, want nil", s.Second)
	}
}

func TestParseSecondsRejectsBadInput(t *testing.T) {
	cases := []string{
		"60 * * * * *",  // second out of range
		"* * * * * * *", // too many fields
		"x * * * * *",   // not a number
		"5-1 * * * * *", // descending second range
	}
	for _, c := range cases {
		if _, err := Parse(c); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", c)
		}
	}
}

func TestSecondsRoundTrip(t *testing.T) {
	s, err := Parse("5,10-20 */15 9-17 * * 1-5")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	s2, err := Parse(s.String())
	if err != nil {
		t.Fatalf("Parse(%q) error: %v", s.String(), err)
	}
	if s.String() != s2.String() {
		t.Errorf("round trip mismatch: %q -> %q", s.String(), s2.String())
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
