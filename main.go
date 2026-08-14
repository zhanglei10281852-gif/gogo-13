package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ops/cronchecker/pkg/output"
	"github.com/ops/cronchecker/pkg/task"
	"github.com/ops/cronchecker/pkg/timeline"
)

func main() {
	var (
		inputFile    string
		windowDays   int
		windowHours  int
		outFormat    string
		strict       bool
		spikeWindow  string
		spikeThresh  int
		suggestLimit int
	)

	flag.StringVar(&inputFile, "f", "", "任务清单文件（.json 或 .csv），必填")
	flag.IntVar(&windowDays, "days", 7, "向前推演的天数")
	flag.IntVar(&windowHours, "hours", 0, "向前推演的小时数（与 days 叠加）")
	flag.StringVar(&outFormat, "format", "text", "输出格式: text 或 json")
	flag.BoolVar(&strict, "strict", false, "检测到冲突时以非零退出码退出")
	flag.StringVar(&spikeWindow, "spike-window", "5m", "惊群检测时间窗口（如 5m, 30s, 1h）")
	flag.IntVar(&spikeThresh, "spike-threshold", 3, "惊群检测阈值（窗口内达到多少任务触发算尖峰）")
	flag.IntVar(&suggestLimit, "suggest-shift", 30, "错峰建议最大延迟分钟数")
	flag.Parse()

	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "错误: 必须用 -f 指定任务清单文件")
		flag.Usage()
		os.Exit(2)
	}

	tasks, err := task.LoadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载任务清单失败: %v\n", err)
		os.Exit(2)
	}

	now := time.Now()
	windowStart := now.Truncate(time.Minute)
	windowEnd := windowStart.AddDate(0, 0, windowDays).Add(time.Duration(windowHours) * time.Hour)

	execs := task.ExpandExecutions(tasks, windowStart, windowEnd)
	conflicts := timeline.DetectOverlaps(execs)

	sw, err := time.ParseDuration(spikeWindow)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析 spike-window 失败: %v\n", err)
		os.Exit(2)
	}
	spikes := timeline.DetectSpikes(execs, sw, spikeThresh)
	suggestions := timeline.SuggestReschedule(conflicts, suggestLimit)

	hasConflict := len(conflicts) > 0 || len(spikes) > 0
	report := output.Report{
		WindowStart: windowStart,
		WindowEnd:   windowEnd,
		Tasks:       tasks,
		Executions:  execs,
		Conflicts:   conflicts,
		Spikes:      spikes,
		Suggestions: suggestions,
		HasConflict: hasConflict,
	}

	switch outFormat {
	case "json":
		if err := output.FormatJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "输出 JSON 失败: %v\n", err)
			os.Exit(2)
		}
	case "text":
		if err := output.FormatText(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "输出失败: %v\n", err)
			os.Exit(2)
		}
	default:
		fmt.Fprintf(os.Stderr, "未知格式 %q（支持 text / json）\n", outFormat)
		os.Exit(2)
	}

	if strict && hasConflict {
		os.Exit(1)
	}
}
