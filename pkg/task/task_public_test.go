package task_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ops/cronchecker/pkg/task"
)

const maxDurationInt64 = int64(1<<63 - 1)

func loadDuration(t *testing.T, format, duration string) ([]task.Task, error) {
	t.Helper()

	var content string
	switch format {
	case "json":
		content = fmt.Sprintf(`[{"name":"job","cron":"0 10 * * *","command":"run","duration":%q,"host":"host-a"}]`, duration)
	case "csv":
		content = "name,cron,command,duration,host\n" +
			fmt.Sprintf("job,0 10 * * *,run,%s,host-a\n", duration)
	default:
		t.Fatalf("unknown format %q", format)
	}

	path := filepath.Join(t.TempDir(), "tasks."+format)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return task.LoadFile(path)
}

func TestLoadFileRejectsNonPositiveAndOverflowingDurations(t *testing.T) {
	maxMinutes := maxDurationInt64 / int64(time.Minute)
	maxHours := maxDurationInt64 / int64(time.Hour)
	maxSeconds := maxDurationInt64 / int64(time.Second)

	invalid := []string{
		"0s", "-1s", // accepted by time.ParseDuration
		"0", "-1", "0min", "-1hour",
		fmt.Sprint(maxMinutes + 1),
		fmt.Sprintf("%dmin", maxMinutes+1),
		fmt.Sprintf("%dhour", maxHours+1),
		fmt.Sprintf("%dsec", maxSeconds+1),
	}

	for _, format := range []string{"json", "csv"} {
		for _, duration := range invalid {
			t.Run(format+"/"+duration, func(t *testing.T) {
				if tasks, err := loadDuration(t, format, duration); err == nil {
					t.Fatalf("LoadFile accepted duration %q: %+v", duration, tasks)
				}
			})
		}
	}
}

func TestLoadFileAndExpandExecutionsAcceptValidDurations(t *testing.T) {
	maxMinutes := maxDurationInt64 / int64(time.Minute)
	maxHours := maxDurationInt64 / int64(time.Hour)
	maxSeconds := maxDurationInt64 / int64(time.Second)

	valid := []struct {
		input string
		want  time.Duration
	}{
		{"1ns", time.Nanosecond},
		{"1m30s", 90 * time.Second},
		{"15", 15 * time.Minute},
		{fmt.Sprint(maxMinutes), time.Duration(maxMinutes) * time.Minute},
		{"5min", 5 * time.Minute},
		{fmt.Sprintf("%dmin", maxMinutes), time.Duration(maxMinutes) * time.Minute},
		{"2hour", 2 * time.Hour},
		{fmt.Sprintf("%dhour", maxHours), time.Duration(maxHours) * time.Hour},
		{"3sec", 3 * time.Second},
		{fmt.Sprintf("%dsec", maxSeconds), time.Duration(maxSeconds) * time.Second},
	}
	start := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	for _, format := range []string{"json", "csv"} {
		for _, tc := range valid {
			t.Run(format+"/"+tc.input, func(t *testing.T) {
				tasks, err := loadDuration(t, format, tc.input)
				if err != nil {
					t.Fatalf("LoadFile(%q): %v", tc.input, err)
				}
				if len(tasks) != 1 || tasks[0].Duration != tc.want {
					t.Fatalf("duration %q: got %+v, want %v", tc.input, tasks, tc.want)
				}

				executions := task.ExpandExecutions(tasks, start, end)
				if len(executions) != 1 {
					t.Fatalf("ExpandExecutions returned %d executions, want 1", len(executions))
				}
				execution := executions[0]
				if !execution.End.After(execution.Start) {
					t.Fatalf("duration %q produced non-positive execution: %v to %v", tc.input, execution.Start, execution.End)
				}
				if got := execution.End.Sub(execution.Start); got != tc.want {
					t.Fatalf("execution duration = %v, want %v", got, tc.want)
				}
			})
		}
	}
}
