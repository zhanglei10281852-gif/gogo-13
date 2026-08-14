package timeline

import (
	"sort"
	"time"

	"github.com/ops/cronchecker/pkg/task"
)

type OverlapConflict struct {
	ScopeKey   string
	ScopeType  string
	A          task.Execution
	B          task.Execution
	OverlapStart time.Time
	OverlapEnd   time.Time
}

func overlaps(a, b task.Execution) (bool, time.Time, time.Time) {
	start := a.Start
	if b.Start.After(start) {
		start = b.Start
	}
	end := a.End
	if b.End.Before(end) {
		end = b.End
	}
	if start.Before(end) {
		return true, start, end
	}
	return false, time.Time{}, time.Time{}
}

func groupByScope(exs []task.Execution) map[string][]task.Execution {
	groups := make(map[string][]task.Execution)
	for _, ex := range exs {
		if ex.Host != "" {
			key := "host:" + ex.Host
			groups[key] = append(groups[key], ex)
		}
		if ex.Resource != "" {
			key := "res:" + ex.Resource
			groups[key] = append(groups[key], ex)
		}
	}
	return groups
}

func DetectOverlaps(exs []task.Execution) []OverlapConflict {
	var conflicts []OverlapConflict
	groups := groupByScope(exs)

	for scopeKey, group := range groups {
		scopeType := "host"
		if len(scopeKey) > 4 && scopeKey[:4] == "res:" {
			scopeType = "resource"
		}

		sort.Slice(group, func(i, j int) bool {
			return group[i].Start.Before(group[j].Start)
		})

		for i := 0; i < len(group); i++ {
			for j := i + 1; j < len(group); j++ {
				if group[j].Start.After(group[i].End) || group[j].Start.Equal(group[i].End) {
					break
				}
				ok, oStart, oEnd := overlaps(group[i], group[j])
				if ok {
					if group[i].Task.Name == group[j].Task.Name &&
						group[i].Start.Equal(group[j].Start) {
						continue
					}
					conflicts = append(conflicts, OverlapConflict{
						ScopeKey:     scopeKey,
						ScopeType:    scopeType,
						A:            group[i],
						B:            group[j],
						OverlapStart: oStart,
						OverlapEnd:   oEnd,
					})
				}
			}
		}
	}
	return conflicts
}

func GroupConflictsByScope(conflicts []OverlapConflict) map[string][]OverlapConflict {
	m := make(map[string][]OverlapConflict)
	for _, c := range conflicts {
		m[c.ScopeKey] = append(m[c.ScopeKey], c)
	}
	return m
}
