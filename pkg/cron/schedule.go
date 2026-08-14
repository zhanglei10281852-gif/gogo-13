package cron

import (
	"time"
)

func (s *Schedule) Next(after time.Time) (time.Time, bool) {
	t := after.Truncate(time.Minute).Add(time.Minute)
	return s.nextFrom(t)
}

func (s *Schedule) NextInclusive(at time.Time) (time.Time, bool) {
	t := at.Truncate(time.Minute)
	if s.Matches(t) {
		return t, true
	}
	return s.nextFrom(t.Add(time.Minute))
}

func (s *Schedule) nextFrom(start time.Time) (time.Time, bool) {
	const maxIterations = 366 * 24 * 60
	t := start
	for i := 0; i < maxIterations; i++ {
		if s.Matches(t) {
			return t, true
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}, false
}

func (s *Schedule) AllBetween(start, end time.Time) []time.Time {
	var results []time.Time
	start = start.Truncate(time.Minute)
	end = end.Truncate(time.Minute)

	t := start
	if !s.Matches(t) {
		next, ok := s.NextInclusive(t)
		if !ok {
			return results
		}
		t = next
	}

	for !t.After(end) {
		results = append(results, t)
		next, ok := s.Next(t)
		if !ok {
			break
		}
		t = next
	}
	return results
}
