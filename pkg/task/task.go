package task

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ops/cronchecker/pkg/cron"
)

type Task struct {
	Name     string        `json:"name"`
	CronExpr string        `json:"cron"`
	Command  string        `json:"command"`
	Duration time.Duration `json:"duration"`
	Host     string        `json:"host"`
	Resource string        `json:"resource"`
	Schedule *cron.Schedule `json:"-"`
}

type rawTaskJSON struct {
	Name     string `json:"name"`
	CronExpr string `json:"cron"`
	Command  string `json:"command"`
	Duration string `json:"duration"`
	Host     string `json:"host"`
	Resource string `json:"resource"`
}

type rawTaskCSV struct {
	Name     string
	CronExpr string
	Command  string
	Duration string
	Host     string
	Resource string
}

func (t *Task) parseSchedule() error {
	s, err := cron.Parse(t.CronExpr)
	if err != nil {
		return fmt.Errorf("task %q: invalid cron %q: %w", t.Name, t.CronExpr, err)
	}
	t.Schedule = s
	return nil
}

type durationSyntax struct {
	suffix string
	unit   time.Duration
}

var customDurationSyntaxes = []durationSyntax{
	{suffix: "min", unit: time.Minute},
	{suffix: "m", unit: time.Minute},
	{suffix: "hour", unit: time.Hour},
	{suffix: "hr", unit: time.Hour},
	{suffix: "h", unit: time.Hour},
	{suffix: "sec", unit: time.Second},
	{suffix: "s", unit: time.Second},
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if d, err := time.ParseDuration(s); err == nil {
		return validatePositiveDuration(s, d)
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return scaleDuration(s, n, time.Minute)
	}
	for _, syntax := range customDurationSyntaxes {
		if !strings.HasSuffix(s, syntax.suffix) {
			continue
		}
		value := strings.TrimSpace(strings.TrimSuffix(s, syntax.suffix))
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse duration %q: %w", s, err)
		}
		return scaleDuration(s, n, syntax.unit)
	}
	return 0, fmt.Errorf("cannot parse duration %q", s)
}

func validatePositiveDuration(input string, d time.Duration) (time.Duration, error) {
	if d <= 0 {
		return 0, fmt.Errorf("duration %q must be positive", input)
	}
	return d, nil
}

func scaleDuration(input string, value int64, unit time.Duration) (time.Duration, error) {
	if value <= 0 {
		return 0, fmt.Errorf("duration %q must be positive", input)
	}
	const maxDuration = int64(1<<63 - 1)
	if value > maxDuration/int64(unit) {
		return 0, fmt.Errorf("duration %q overflows time.Duration", input)
	}
	return validatePositiveDuration(input, time.Duration(value)*unit)
}

func LoadFile(path string) ([]Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return loadJSON(f)
	case ".csv":
		return loadCSV(f)
	default:
		return nil, fmt.Errorf("unsupported file extension %q (use .json or .csv)", ext)
	}
}

func loadJSON(r io.Reader) ([]Task, error) {
	var raw []rawTaskJSON
	dec := json.NewDecoder(r)
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	tasks := make([]Task, 0, len(raw))
	for i, rt := range raw {
		d, err := parseDuration(rt.Duration)
		if err != nil {
			return nil, fmt.Errorf("task %d (%q): %w", i, rt.Name, err)
		}
		t := Task{
			Name:     rt.Name,
			CronExpr: rt.CronExpr,
			Command:  rt.Command,
			Duration: d,
			Host:     rt.Host,
			Resource: rt.Resource,
		}
		if err := t.parseSchedule(); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func loadCSV(r io.Reader) ([]Task, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv read: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("csv: need at least header + 1 data row, got %d", len(records))
	}

	header := records[0]
	idx := map[string]int{}
	for i, h := range header {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	required := []string{"name", "cron", "command", "duration", "host"}
	for _, r := range required {
		if _, ok := idx[r]; !ok {
			return nil, fmt.Errorf("csv missing required column %q", r)
		}
	}

	tasks := make([]Task, 0, len(records)-1)
	for i, rec := range records[1:] {
		get := func(key string) string {
			if j, ok := idx[key]; ok && j < len(rec) {
				return strings.TrimSpace(rec[j])
			}
			return ""
		}
		d, err := parseDuration(get("duration"))
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		t := Task{
			Name:     get("name"),
			CronExpr: get("cron"),
			Command:  get("command"),
			Duration: d,
			Host:     get("host"),
			Resource: get("resource"),
		}
		if t.Name == "" {
			t.Name = fmt.Sprintf("task-%d", i+1)
		}
		if err := t.parseSchedule(); err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

type Execution struct {
	Task      Task
	Start     time.Time
	End       time.Time
	Host      string
	Resource  string
}

func ExpandExecutions(tasks []Task, start, end time.Time) []Execution {
	var exs []Execution
	for _, t := range tasks {
		fires := t.Schedule.AllBetween(start, end)
		for _, f := range fires {
			exs = append(exs, Execution{
				Task:     t,
				Start:    f,
				End:      f.Add(t.Duration),
				Host:     t.Host,
				Resource: t.Resource,
			})
		}
	}
	return exs
}
