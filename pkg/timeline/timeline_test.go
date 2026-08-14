package timeline

import (
	"testing"
	"time"

	"github.com/ops/cronchecker/pkg/cron"
	"github.com/ops/cronchecker/pkg/task"
)

func mustTask(name, expr, host string, dur time.Duration) task.Task {
	s, _ := cron.Parse(expr)
	return task.Task{Name: name, CronExpr: expr, Host: host, Duration: dur, Schedule: s}
}

func TestDetectOverlaps_SimpleOverlap(t *testing.T) {
	start := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 11, 0, 0, 0, time.UTC)

	tasks := []task.Task{
		mustTask("A", "0 10 * * *", "host1", 30*time.Minute),
		mustTask("B", "15 10 * * *", "host1", 30*time.Minute),
	}
	exs := task.ExpandExecutions(tasks, start, end)
	conflicts := DetectOverlaps(exs)
	if len(conflicts) != 1 {
		t.Fatalf("want 1 conflict, got %d", len(conflicts))
	}
	c := conflicts[0]
	if c.OverlapStart.IsZero() || c.OverlapEnd.IsZero() {
		t.Errorf("overlap times should be set")
	}
	if c.OverlapStart.After(c.OverlapEnd) {
		t.Errorf("start after end")
	}
	names := map[string]bool{c.A.Task.Name: true, c.B.Task.Name: true}
	if !names["A"] || !names["B"] {
		t.Errorf("conflict should involve A and B, got %v-%v", c.A.Task.Name, c.B.Task.Name)
	}
}

func TestDetectOverlaps_NoOverlap(t *testing.T) {
	start := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tasks := []task.Task{
		mustTask("A", "0 10 * * *", "host1", 30*time.Minute),
		mustTask("B", "0 11 * * *", "host1", 30*time.Minute),
	}
	exs := task.ExpandExecutions(tasks, start, end)
	conflicts := DetectOverlaps(exs)
	if len(conflicts) != 0 {
		t.Errorf("A ends at 10:30, B starts at 11:00, no overlap, but got %d", len(conflicts))
	}
}

func TestDetectOverlaps_Adjacent(t *testing.T) {
	start := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	tasks := []task.Task{
		mustTask("A", "0 10 * * *", "host1", 60*time.Minute),
		mustTask("B", "0 11 * * *", "host1", 30*time.Minute),
	}
	exs := task.ExpandExecutions(tasks, start, end)
	conflicts := DetectOverlaps(exs)
	if len(conflicts) != 0 {
		t.Errorf("A ends exactly when B starts (11:00), should be no overlap, got %d", len(conflicts))
	}
}

func TestDetectOverlaps_DifferentHosts(t *testing.T) {
	start := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 11, 0, 0, 0, time.UTC)

	tasks := []task.Task{
		mustTask("A", "0 10 * * *", "host1", 30*time.Minute),
		mustTask("B", "15 10 * * *", "host2", 30*time.Minute),
	}
	exs := task.ExpandExecutions(tasks, start, end)
	conflicts := DetectOverlaps(exs)
	if len(conflicts) != 0 {
		t.Errorf("different hosts should not conflict, got %d", len(conflicts))
	}
}

func TestDetectOverlaps_SharedResource(t *testing.T) {
	start := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 11, 0, 0, 0, time.UTC)

	t1 := mustTask("A", "0 10 * * *", "host1", 30*time.Minute)
	t1.Resource = "db"
	t2 := mustTask("B", "15 10 * * *", "host2", 30*time.Minute)
	t2.Resource = "db"
	exs := task.ExpandExecutions([]task.Task{t1, t2}, start, end)
	conflicts := DetectOverlaps(exs)
	foundResConflict := false
	for _, c := range conflicts {
		if c.ScopeType == "resource" {
			foundResConflict = true
		}
	}
	if !foundResConflict {
		t.Errorf("shared resource should cause conflict")
	}
}

func TestDetectSpikes_ThunderingHerd(t *testing.T) {
	start := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 1, 0, 0, 0, time.UTC)

	var tasks []task.Task
	for i := 0; i < 5; i++ {
		tasks = append(tasks, mustTask("job-"+string(rune('A'+i)), "0 0 * * *", "host1", 5*time.Minute))
	}
	exs := task.ExpandExecutions(tasks, start, end)
	spikes := DetectSpikes(exs, 5*time.Minute, 3)
	if len(spikes) == 0 {
		t.Errorf("5 tasks firing at midnight should be a spike")
	}
	if spikes[0].Count < 5 {
		t.Errorf("spike count should be 5, got %d", spikes[0].Count)
	}
}

func TestDetectSpikes_NoSpike(t *testing.T) {
	start := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 2, 0, 0, 0, time.UTC)

	tasks := []task.Task{
		mustTask("A", "0 0 * * *", "host1", 5*time.Minute),
		mustTask("B", "30 0 * * *", "host1", 5*time.Minute),
	}
	exs := task.ExpandExecutions(tasks, start, end)
	spikes := DetectSpikes(exs, 5*time.Minute, 3)
	if len(spikes) != 0 {
		t.Errorf("2 tasks spread across an hour should not be a spike, got %d", len(spikes))
	}
}

func TestSuggestReschedule_Simple(t *testing.T) {
	start := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 15, 11, 0, 0, 0, time.UTC)
	tasks := []task.Task{
		mustTask("A", "0 10 * * *", "host1", 30*time.Minute),
		mustTask("B", "15 10 * * *", "host1", 30*time.Minute),
	}
	exs := task.ExpandExecutions(tasks, start, end)
	conflicts := DetectOverlaps(exs)
	suggests := SuggestReschedule(conflicts, 60)
	if len(suggests) != 1 {
		t.Fatalf("want 1 suggestion, got %d", len(suggests))
	}
	if suggests[0].ShiftMinutes <= 0 {
		t.Errorf("shift must be positive, got %d", suggests[0].ShiftMinutes)
	}
}
