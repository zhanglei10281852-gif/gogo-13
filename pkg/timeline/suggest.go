package timeline

import (
	"fmt"
)

type RescheduleSuggestion struct {
	Conflict   OverlapConflict
	ShiftTask  string
	ShiftMinutes int
	Reason     string
}

func SuggestReschedule(conflicts []OverlapConflict, maxShiftMinutes int) []RescheduleSuggestion {
	if maxShiftMinutes <= 0 {
		maxShiftMinutes = 30
	}
	var suggestions []RescheduleSuggestion

	for _, c := range conflicts {
		aDur := c.A.Task.Duration
		bDur := c.B.Task.Duration

		aEnd := c.A.Start.Add(aDur)
		neededAfterA := c.B.Start.Sub(aEnd)
		shiftB := int(-neededAfterA.Minutes()) + 1
		if shiftB > 0 && shiftB <= maxShiftMinutes {
			suggestions = append(suggestions, RescheduleSuggestion{
				Conflict:     c,
				ShiftTask:    c.B.Task.Name,
				ShiftMinutes: shiftB,
				Reason: fmt.Sprintf("往后挪 %d 分钟可在 %s 结束后启动，避开与 %s 重叠",
					shiftB, c.A.Task.Name, c.A.Task.Name),
			})
			continue
		}

		bEnd := c.B.Start.Add(bDur)
		neededAfterB := c.A.Start.Sub(bEnd)
		shiftA := int(-neededAfterB.Minutes()) + 1
		if shiftA > 0 && shiftA <= maxShiftMinutes {
			suggestions = append(suggestions, RescheduleSuggestion{
				Conflict:     c,
				ShiftTask:    c.A.Task.Name,
				ShiftMinutes: shiftA,
				Reason: fmt.Sprintf("往后挪 %d 分钟可在 %s 结束后启动，避开与 %s 重叠",
					shiftA, c.B.Task.Name, c.B.Task.Name),
			})
			continue
		}

		minShift := shiftB
		if shiftA < minShift {
			minShift = shiftA
		}
		if minShift > 0 {
			target := c.A.Task.Name
			if shiftB < shiftA {
				target = c.B.Task.Name
				minShift = shiftB
			} else {
				minShift = shiftA
			}
			suggestions = append(suggestions, RescheduleSuggestion{
				Conflict:     c,
				ShiftTask:    target,
				ShiftMinutes: minShift,
				Reason: fmt.Sprintf("至少需要往后挪 %d 分钟才能错开（已超过默认 %d 分钟上限，请人工确认）",
					minShift, maxShiftMinutes),
			})
		}
	}
	return suggestions
}
