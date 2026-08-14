package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type FieldType int

const (
	FieldMinute FieldType = iota
	FieldHour
	FieldDayOfMonth
	FieldMonth
	FieldDayOfWeek
)

var fieldRanges = map[FieldType][2]int{
	FieldMinute:     {0, 59},
	FieldHour:       {0, 23},
	FieldDayOfMonth: {1, 31},
	FieldMonth:      {1, 12},
	FieldDayOfWeek:  {0, 6},
}

var monthNames = map[string]int{
	"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
	"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
}

var weekdayNames = map[string]int{
	"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
}

type Field struct {
	Type     FieldType
	Values   map[int]bool
	Wildcard bool
}

type Schedule struct {
	Minute     Field
	Hour       Field
	DayOfMonth Field
	Month      Field
	DayOfWeek  Field
	RawExpr    string
}

func newField(ft FieldType) Field {
	return Field{
		Type:     ft,
		Values:   make(map[int]bool),
		Wildcard: false,
	}
}

func (f Field) Contains(v int) bool {
	if f.Wildcard {
		return true
	}
	return f.Values[v]
}

func (f Field) IsWildcard() bool {
	return f.Wildcard
}

func Parse(expr string) (*Schedule, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron expression must have 5 fields, got %d: %q", len(fields), expr)
	}

	s := &Schedule{RawExpr: expr}
	types := []FieldType{FieldMinute, FieldHour, FieldDayOfMonth, FieldMonth, FieldDayOfWeek}
	for i, ft := range types {
		f, err := parseField(fields[i], ft)
		if err != nil {
			return nil, fmt.Errorf("field %d (%v): %w", i, ft, err)
		}
		switch ft {
		case FieldMinute:
			s.Minute = f
		case FieldHour:
			s.Hour = f
		case FieldDayOfMonth:
			s.DayOfMonth = f
		case FieldMonth:
			s.Month = f
		case FieldDayOfWeek:
			s.DayOfWeek = f
		}
	}
	return s, nil
}

func parseField(s string, ft FieldType) (Field, error) {
	f := newField(ft)
	rng := fieldRanges[ft]
	minVal, maxVal := rng[0], rng[1]

	s = strings.TrimSpace(s)
	if s == "*" {
		f.Wildcard = true
		return f, nil
	}

	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return f, fmt.Errorf("empty segment in %q", s)
		}
		if err := parseSegment(part, ft, minVal, maxVal, f.Values); err != nil {
			return f, err
		}
	}
	return f, nil
}

func parseSegment(seg string, ft FieldType, minVal, maxVal int, values map[int]bool) error {
	step := 1
	hasStep := false

	if idx := strings.Index(seg, "/"); idx != -1 {
		stepStr := seg[idx+1:]
		var err error
		step, err = strconv.Atoi(stepStr)
		if err != nil {
			return fmt.Errorf("invalid step %q: %w", stepStr, err)
		}
		if step <= 0 {
			return fmt.Errorf("step must be positive, got %d", step)
		}
		seg = seg[:idx]
		hasStep = true
	}

	var start, end int
	if seg == "*" {
		start = minVal
		end = maxVal
	} else if idx := strings.Index(seg, "-"); idx != -1 {
		s1, err := parseValue(seg[:idx], ft)
		if err != nil {
			return err
		}
		s2, err := parseValue(seg[idx+1:], ft)
		if err != nil {
			return err
		}
		start = s1
		end = s2
	} else {
		v, err := parseValue(seg, ft)
		if err != nil {
			return err
		}
		if !hasStep {
			if v < minVal || v > maxVal {
				return fmt.Errorf("value %d out of range [%d, %d]", v, minVal, maxVal)
			}
			values[v] = true
			return nil
		}
		start = v
		end = maxVal
	}

	if start < minVal || start > maxVal {
		return fmt.Errorf("range start %d out of range [%d, %d]", start, minVal, maxVal)
	}
	if end < minVal || end > maxVal {
		return fmt.Errorf("range end %d out of range [%d, %d]", end, minVal, maxVal)
	}
	if start > end {
		return fmt.Errorf("range start %d > end %d", start, end)
	}

	for v := start; v <= end; v += step {
		values[v] = true
	}
	return nil
}

func parseValue(s string, ft FieldType) (int, error) {
	s = strings.TrimSpace(s)
	if ft == FieldMonth {
		if v, ok := monthNames[strings.ToUpper(s)]; ok {
			return v, nil
		}
	}
	if ft == FieldDayOfWeek {
		if v, ok := weekdayNames[strings.ToUpper(s)]; ok {
			return v, nil
		}
	}
	return strconv.Atoi(s)
}

func (s *Schedule) dayMatchEnabled() (bool, bool) {
	domSet := !s.DayOfMonth.IsWildcard()
	dowSet := !s.DayOfWeek.IsWildcard()
	return domSet, dowSet
}

func (s *Schedule) Matches(t time.Time) bool {
	if !s.Month.Contains(int(t.Month())) {
		return false
	}
	if !s.Hour.Contains(t.Hour()) {
		return false
	}
	if !s.Minute.Contains(t.Minute()) {
		return false
	}

	domSet, dowSet := s.dayMatchEnabled()
	domOK := s.DayOfMonth.Contains(t.Day())
	dowOK := s.DayOfWeek.Contains(int(t.Weekday()))

	if domSet && dowSet {
		return domOK || dowOK
	}
	if domSet {
		return domOK
	}
	if dowSet {
		return dowOK
	}
	return domOK
}
