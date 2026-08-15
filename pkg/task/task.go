package task

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
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

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// time.ParseDuration understands composite durations (e.g. "1m30s") and
	// rejects overflow on its own, but it also accepts zero/negative values,
	// which we must reject so that End stays strictly after Start.
	if d, err := time.ParseDuration(s); err == nil {
		if d <= 0 {
			return 0, fmt.Errorf("duration must be greater than zero, got %q", s)
		}
		return d, nil
	}

	// Plain integer (e.g. "15") => minutes.
	if n, err := strconv.Atoi(s); err == nil {
		return mulUnit(n, time.Minute, s)
	}

	// Suffix forms not understood by time.ParseDuration (min/hr/hour/sec).
	if n, ok := parseIntWithSuffix(s, "min", "m"); ok {
		return mulUnit(n, time.Minute, s)
	}
	if n, ok := parseIntWithSuffix(s, "hour", "hr", "h"); ok {
		return mulUnit(n, time.Hour, s)
	}
	if n, ok := parseIntWithSuffix(s, "sec", "s"); ok {
		return mulUnit(n, time.Second, s)
	}
	return 0, fmt.Errorf("cannot parse duration %q", s)
}

// parseIntWithSuffix reports whether s ends with any of suffixes; if so it
// strips all matching suffixes and parses the remainder as an integer.
func parseIntWithSuffix(s string, suffixes ...string) (int, bool) {
	matched := false
	for _, suf := range suffixes {
		if strings.HasSuffix(s, suf) {
			matched = true
			s = strings.TrimSuffix(s, suf)
		}
	}
	if !matched {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, false
	}
	return n, true
}

// mulUnit scales n by unit, rejecting non-positive magnitudes and products
// that would overflow time.Duration (which otherwise wrap to negative values
// and silently break conflict detection).
func mulUnit(n int, unit time.Duration, original string) (time.Duration, error) {
	if n <= 0 {
		return 0, fmt.Errorf("duration must be greater than zero, got %q", original)
	}
	if int64(n) > math.MaxInt64/int64(unit) {
		return 0, fmt.Errorf("duration %q overflows time.Duration", original)
	}
	return time.Duration(n) * unit, nil
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
