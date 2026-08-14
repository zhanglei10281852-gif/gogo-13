package cron

import (
	"testing"
	"time"
)

func TestParseBasic(t *testing.T) {
	cases := []struct {
		expr string
		ok   bool
	}{
		{"* * * * *", true},
		{"0 0 * * *", true},
		{"*/5 * * * *", true},
		{"1,15,30 * * * *", true},
		{"1-5 * * * *", true},
		{"0 0 1 1 *", true},
		{"0 0 * * MON", true},
		{"0 0 1 JAN *", true},
		{"* * *", false},
		{"* * * * * *", false},
		{"60 * * * *", false},
		{"* 24 * * *", false},
		{"* * 0 * *", false},
		{"* * 32 * *", false},
		{"* * * 0 *", false},
		{"* * * 13 *", false},
		{"* * * * 7", false},
	}
	for _, c := range cases {
		_, err := Parse(c.expr)
		if (err == nil) != c.ok {
			t.Errorf("Parse(%q) ok=%v want ok=%v err=%v", c.expr, err == nil, c.ok, err)
		}
	}
}

func TestFieldValues(t *testing.T) {
	cases := []struct {
		expr   string
		ft     FieldType
		expect []int
	}{
		{"*/15 * * * *", FieldMinute, []int{0, 15, 30, 45}},
		{"1-5 * * * *", FieldMinute, []int{1, 2, 3, 4, 5}},
		{"1,15,30 * * * *", FieldMinute, []int{1, 15, 30}},
		{"10-20/2 * * * *", FieldMinute, []int{10, 12, 14, 16, 18, 20}},
		{"* 9-17 * * *", FieldHour, []int{9, 10, 11, 12, 13, 14, 15, 16, 17}},
	}
	for _, c := range cases {
		s, err := Parse(c.expr)
		if err != nil {
			t.Fatalf("Parse(%q): %v", c.expr, err)
		}
		var f Field
		switch c.ft {
		case FieldMinute:
			f = s.Minute
		case FieldHour:
			f = s.Hour
		}
		for _, v := range c.expect {
			if !f.Contains(v) {
				t.Errorf("%q field %v should contain %d", c.expr, c.ft, v)
			}
		}
	}
}

func TestMatchesWildcard(t *testing.T) {
	s, _ := Parse("* * * * *")
	if !s.Matches(time.Date(2024, 2, 29, 12, 34, 0, 0, time.UTC)) {
		t.Error("wildcard should match everything")
	}
}

func TestMatchesSpecificMinute(t *testing.T) {
	s, _ := Parse("30 9 * * *")
	ok1 := s.Matches(time.Date(2024, 6, 15, 9, 30, 0, 0, time.UTC))
	ok2 := s.Matches(time.Date(2024, 6, 15, 9, 31, 0, 0, time.UTC))
	ok3 := s.Matches(time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC))
	if !ok1 {
		t.Error("should match 9:30")
	}
	if ok2 {
		t.Error("should not match 9:31")
	}
	if ok3 {
		t.Error("should not match 10:30")
	}
}

func TestMatchesDayOfMonthAndWeekOrSemantics(t *testing.T) {
	s, _ := Parse("0 12 15 * MON")
	june15 := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	if june15.Weekday() != time.Saturday {
		t.Fatalf("test assumption wrong: 2024-06-15 is %v", june15.Weekday())
	}
	if !s.Matches(june15) {
		t.Error("day-of-month 15 should match regardless of weekday (OR semantics)")
	}

	mon := time.Date(2024, 6, 17, 12, 0, 0, 0, time.UTC)
	if mon.Weekday() != time.Monday {
		t.Fatalf("test assumption wrong: 2024-06-17 is %v", mon.Weekday())
	}
	if !s.Matches(mon) {
		t.Error("Monday should match regardless of day-of-month (OR semantics)")
	}

	wed := time.Date(2024, 6, 19, 12, 0, 0, 0, time.UTC)
	if wed.Weekday() != time.Wednesday {
		t.Fatalf("test assumption wrong")
	}
	if wed.Day() == 15 {
		t.Fatalf("test assumption wrong")
	}
	if s.Matches(wed) {
		t.Error("Wednesday the 19th should NOT match (neither 15th nor Monday)")
	}
}

func TestMatchesFebruaryLeapYear(t *testing.T) {
	s, _ := Parse("0 0 29 2 *")
	leap := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	if !s.Matches(leap) {
		t.Error("should match Feb 29 on leap year")
	}
	nonleap := time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC)
	if s.Matches(nonleap) {
		t.Error("should not match Feb 28 when 29 is specified")
	}
}

func TestMatchesMonthEnd(t *testing.T) {
	s, _ := Parse("0 0 31 * *")
	jan31 := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	if !s.Matches(jan31) {
		t.Error("should match Jan 31")
	}
	apr30 := time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC)
	if s.Matches(apr30) {
		t.Error("should not match Apr 30 when 31 specified")
	}
}

func TestStepRange(t *testing.T) {
	s, _ := Parse("*/10 * * * *")
	for _, m := range []int{0, 10, 20, 30, 40, 50} {
		if !s.Minute.Contains(m) {
			t.Errorf("*/10 should contain %d", m)
		}
	}
	for _, m := range []int{1, 5, 55} {
		if s.Minute.Contains(m) {
			t.Errorf("*/10 should NOT contain %d", m)
		}
	}
}
