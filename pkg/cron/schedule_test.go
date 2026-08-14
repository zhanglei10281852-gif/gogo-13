package cron

import (
	"testing"
	"time"
)

func TestNextEveryMinute(t *testing.T) {
	s, _ := Parse("* * * * *")
	base := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	next, ok := s.Next(base)
	if !ok {
		t.Fatal("should find next")
	}
	expected := base.Add(time.Minute)
	if !next.Equal(expected) {
		t.Errorf("got %v want %v", next, expected)
	}
}

func TestNextSpecificMinute(t *testing.T) {
	s, _ := Parse("0 * * * *")
	base := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	next, ok := s.Next(base)
	if !ok {
		t.Fatal("should find next")
	}
	expected := time.Date(2024, 6, 15, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("got %v want %v", next, expected)
	}
}

func TestNextInclusive(t *testing.T) {
	s, _ := Parse("30 10 * * *")
	base := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	next, ok := s.NextInclusive(base)
	if !ok {
		t.Fatal("should find")
	}
	if !next.Equal(base) {
		t.Errorf("NextInclusive at matching time should return itself, got %v", next)
	}
}

func TestAllBetweenEvery5Min(t *testing.T) {
	s, _ := Parse("*/5 * * * *")
	start := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 11, 0, 0, 0, time.UTC)
	ts := s.AllBetween(start, end)
	expectedCount := 13 // 0,5,...,60 inclusive
	if len(ts) != expectedCount {
		t.Errorf("got %d want %d, times=%v", len(ts), expectedCount, ts)
	}
	if !ts[0].Equal(start) {
		t.Errorf("first should be start, got %v", ts[0])
	}
	if !ts[len(ts)-1].Equal(end) {
		t.Errorf("last should be end, got %v", ts[len(ts)-1])
	}
}

func TestAllBetweenDailyMidnight(t *testing.T) {
	s, _ := Parse("0 0 * * *")
	start := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC)
	ts := s.AllBetween(start, end)
	if len(ts) != 3 {
		t.Errorf("want 3 days, got %d: %v", len(ts), ts)
	}
}

func TestAllBetweenWeekday(t *testing.T) {
	s, _ := Parse("0 9 * * MON-FRI")
	start := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 21, 23, 59, 0, 0, time.UTC)
	ts := s.AllBetween(start, end)
	expected := 5
	if len(ts) != expected {
		t.Errorf("Mon-Fri in 7 days should give %d, got %d: %v", expected, len(ts), ts)
	}
	for _, tt := range ts {
		if tt.Hour() != 9 || tt.Minute() != 0 {
			t.Errorf("unexpected time %v", tt)
		}
		wd := tt.Weekday()
		if wd == time.Saturday || wd == time.Sunday {
			t.Errorf("weekend should not match, got %v (%v)", tt, wd)
		}
	}
}

func TestAllBetweenEmptyWindow(t *testing.T) {
	s, _ := Parse("0 0 29 2 *")
	start := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 2, 28, 23, 59, 0, 0, time.UTC)
	ts := s.AllBetween(start, end)
	if len(ts) != 0 {
		t.Errorf("non-leap year Feb has no 29th, got %v", ts)
	}
}
