package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/ops/cronchecker/pkg/task"
	"github.com/ops/cronchecker/pkg/timeline"
)

type Report struct {
	WindowStart time.Time
	WindowEnd   time.Time
	Tasks       []task.Task
	Executions  []task.Execution
	Conflicts   []timeline.OverlapConflict
	Spikes      []timeline.SpikeEvent
	Suggestions []timeline.RescheduleSuggestion
	HasConflict bool
}

type execJSON struct {
	TaskName string    `json:"task"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Host     string    `json:"host,omitempty"`
	Resource string    `json:"resource,omitempty"`
}

type conflictJSON struct {
	ScopeType    string    `json:"scope_type"`
	ScopeKey     string    `json:"scope_key"`
	TaskA        string    `json:"task_a"`
	TaskB        string    `json:"task_b"`
	StartA       time.Time `json:"start_a"`
	EndA         time.Time `json:"end_a"`
	StartB       time.Time `json:"start_b"`
	EndB         time.Time `json:"end_b"`
	OverlapStart time.Time `json:"overlap_start"`
	OverlapEnd   time.Time `json:"overlap_end"`
}

type spikeJSON struct {
	Time           time.Time     `json:"time"`
	WindowSeconds  float64       `json:"window_seconds"`
	Count          int           `json:"count"`
	TaskNames      []string      `json:"tasks"`
	HostOrResource string        `json:"scope"`
}

type suggestionJSON struct {
	ScopeKey     string `json:"scope"`
	TaskA        string `json:"task_a"`
	TaskB        string `json:"task_b"`
	ShiftTask    string `json:"shift_task"`
	ShiftMinutes int    `json:"shift_minutes"`
	Reason       string `json:"reason"`
}

type reportJSON struct {
	WindowStart time.Time        `json:"window_start"`
	WindowEnd   time.Time        `json:"window_end"`
	TaskCount   int              `json:"task_count"`
	Tasks       []string         `json:"tasks"`
	Executions  []execJSON       `json:"executions"`
	Conflicts   []conflictJSON   `json:"conflicts"`
	Spikes      []spikeJSON      `json:"spikes"`
	Suggestions []suggestionJSON `json:"suggestions"`
	HasConflict bool             `json:"has_conflict"`
}

func FormatJSON(w io.Writer, r Report) error {
	taskNames := make([]string, 0, len(r.Tasks))
	for _, t := range r.Tasks {
		taskNames = append(taskNames, t.Name)
	}
	execs := make([]execJSON, 0, len(r.Executions))
	for _, e := range r.Executions {
		execs = append(execs, execJSON{
			TaskName: e.Task.Name,
			Start:    e.Start,
			End:      e.End,
			Host:     e.Host,
			Resource: e.Resource,
		})
	}
	conflicts := make([]conflictJSON, 0, len(r.Conflicts))
	for _, c := range r.Conflicts {
		conflicts = append(conflicts, conflictJSON{
			ScopeType:    c.ScopeType,
			ScopeKey:     c.ScopeKey,
			TaskA:        c.A.Task.Name,
			TaskB:        c.B.Task.Name,
			StartA:       c.A.Start,
			EndA:         c.A.End,
			StartB:       c.B.Start,
			EndB:         c.B.End,
			OverlapStart: c.OverlapStart,
			OverlapEnd:   c.OverlapEnd,
		})
	}
	spikes := make([]spikeJSON, 0, len(r.Spikes))
	for _, s := range r.Spikes {
		spikes = append(spikes, spikeJSON{
			Time:           s.Time,
			WindowSeconds:  s.Window.Seconds(),
			Count:          s.Count,
			TaskNames:      s.TaskNames,
			HostOrResource: s.HostOrResource,
		})
	}
	suggestions := make([]suggestionJSON, 0, len(r.Suggestions))
	for _, s := range r.Suggestions {
		suggestions = append(suggestions, suggestionJSON{
			ScopeKey:     s.Conflict.ScopeKey,
			TaskA:        s.Conflict.A.Task.Name,
			TaskB:        s.Conflict.B.Task.Name,
			ShiftTask:    s.ShiftTask,
			ShiftMinutes: s.ShiftMinutes,
			Reason:       s.Reason,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(reportJSON{
		WindowStart: r.WindowStart,
		WindowEnd:   r.WindowEnd,
		TaskCount:   len(r.Tasks),
		Tasks:       taskNames,
		Executions:  execs,
		Conflicts:   conflicts,
		Spikes:      spikes,
		Suggestions: suggestions,
		HasConflict: r.HasConflict,
	})
}

func FormatText(w io.Writer, r Report) error {
	const timeFmt = "2006-01-02 15:04"

	fmt.Fprintf(w, "=== Cron 冲突检测报告 ===\n")
	fmt.Fprintf(w, "时间窗: %s ~ %s\n", r.WindowStart.Format(timeFmt), r.WindowEnd.Format(timeFmt))
	fmt.Fprintf(w, "任务数: %d，总执行次数: %d\n\n", len(r.Tasks), len(r.Executions))

	fmt.Fprintln(w, "--- 任务清单 ---")
	for _, t := range r.Tasks {
		scope := t.Host
		if t.Resource != "" {
			scope += " / res:" + t.Resource
		}
		fmt.Fprintf(w, "  [%s] cron=%q  cmd=%q  时长=%v  scope=%s\n",
			t.Name, t.CronExpr, t.Command, t.Duration, scope)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "--- 时间线（按主机/资源分组）---")
	exByScope := make(map[string][]task.Execution)
	for _, e := range r.Executions {
		if e.Host != "" {
			key := "host:" + e.Host
			exByScope[key] = append(exByScope[key], e)
		}
		if e.Resource != "" {
			key := "res:" + e.Resource
			exByScope[key] = append(exByScope[key], e)
		}
	}
	scopes := make([]string, 0, len(exByScope))
	for k := range exByScope {
		scopes = append(scopes, k)
	}
	sort.Strings(scopes)
	for _, scope := range scopes {
		exs := exByScope[scope]
		sort.Slice(exs, func(i, j int) bool { return exs[i].Start.Before(exs[j].Start) })
		fmt.Fprintf(w, "  [%s] %d 次执行:\n", scope, len(exs))
		for _, e := range exs {
			fmt.Fprintf(w, "    %s - %s  %s\n",
				e.Start.Format(timeFmt), e.End.Format(timeFmt), e.Task.Name)
		}
	}
	fmt.Fprintln(w)

	if len(r.Conflicts) == 0 {
		fmt.Fprintln(w, "--- 冲突检测 ---")
		fmt.Fprintln(w, "  未发现执行区间重叠 ✅")
	} else {
		fmt.Fprintf(w, "--- 冲突检测（%d 处重叠）---\n", len(r.Conflicts))
		byScope := timeline.GroupConflictsByScope(r.Conflicts)
		scopes := make([]string, 0, len(byScope))
		for k := range byScope {
			scopes = append(scopes, k)
		}
		sort.Strings(scopes)
		for _, scope := range scopes {
			fmt.Fprintf(w, "  [%s]:\n", scope)
			for _, c := range byScope[scope] {
				fmt.Fprintf(w, "    🔴 %s (%s~%s) 与  %s (%s~%s)\n",
					c.A.Task.Name, c.A.Start.Format(timeFmt), c.A.End.Format(timeFmt),
					c.B.Task.Name, c.B.Start.Format(timeFmt), c.B.End.Format(timeFmt))
				fmt.Fprintf(w, "       重叠时段: %s ~ %s\n",
					c.OverlapStart.Format(timeFmt), c.OverlapEnd.Format(timeFmt))
			}
		}
	}
	fmt.Fprintln(w)

	if len(r.Spikes) == 0 {
		fmt.Fprintln(w, "--- 惊群尖峰 ---")
		fmt.Fprintln(w, "  未检测到短时间大量任务并发 ✅")
	} else {
		fmt.Fprintf(w, "--- 惊群尖峰（%d 处）---\n", len(r.Spikes))
		for _, s := range r.Spikes {
			fmt.Fprintf(w, "    ⚡ %s [%s] %d 个任务在 %v 窗口内触发: %s\n",
				s.Time.Format(timeFmt), s.HostOrResource, s.Count,
				s.Window, strings.Join(s.TaskNames, ", "))
		}
	}
	fmt.Fprintln(w)

	if len(r.Suggestions) > 0 {
		fmt.Fprintf(w, "--- 错峰建议（%d 条）---\n", len(r.Suggestions))
		for _, s := range r.Suggestions {
			fmt.Fprintf(w, "    💡 [%s] %s 与 %s 冲突: 将 %s %s\n",
				s.Conflict.ScopeKey, s.Conflict.A.Task.Name, s.Conflict.B.Task.Name,
				s.ShiftTask, s.Reason)
		}
		fmt.Fprintln(w)
	}

	if r.HasConflict {
		fmt.Fprintln(w, "⚠️  检测到冲突，使用 --strict 可使进程以非零退出码退出")
	} else {
		fmt.Fprintln(w, "✅ 所有任务排布良好，无冲突")
	}
	return nil
}
