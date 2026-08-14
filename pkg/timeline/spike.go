package timeline

import (
	"sort"
	"time"

	"github.com/ops/cronchecker/pkg/task"
)

type SpikeEvent struct {
	Time        time.Time
	Window      time.Duration
	Count       int
	TaskNames   []string
	HostOrResource string
}

const DefaultSpikeWindow = 5 * time.Minute
const DefaultSpikeThreshold = 3

type SpikeDetector struct {
	Window    time.Duration
	Threshold int
}

func (sd SpikeDetector) defaults() SpikeDetector {
	if sd.Window <= 0 {
		sd.Window = DefaultSpikeWindow
	}
	if sd.Threshold <= 0 {
		sd.Threshold = DefaultSpikeThreshold
	}
	return sd
}

func DetectSpikes(exs []task.Execution, window time.Duration, threshold int) []SpikeEvent {
	sd := SpikeDetector{Window: window, Threshold: threshold}.defaults()

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

	var spikes []SpikeEvent
	for key, group := range groups {
		sort.Slice(group, func(i, j int) bool {
			return group[i].Start.Before(group[j].Start)
		})

		seen := make(map[int64]bool)
		for i := 0; i < len(group); i++ {
			windowEnd := group[i].Start.Add(sd.Window)
			var batch []task.Execution
			for j := i; j < len(group); j++ {
				if group[j].Start.After(windowEnd) {
					break
				}
				batch = append(batch, group[j])
			}
			if len(batch) >= sd.Threshold {
				bucket := group[i].Start.Unix() / int64(sd.Window.Seconds())
				if seen[bucket] {
					continue
				}
				seen[bucket] = true
				names := make([]string, 0, len(batch))
				seenNames := make(map[string]bool)
				for _, e := range batch {
					if !seenNames[e.Task.Name] {
						names = append(names, e.Task.Name)
						seenNames[e.Task.Name] = true
					}
				}
				spikes = append(spikes, SpikeEvent{
					Time:           group[i].Start,
					Window:         sd.Window,
					Count:          len(batch),
					TaskNames:      names,
					HostOrResource: key,
				})
			}
		}
	}

	sort.Slice(spikes, func(i, j int) bool {
		return spikes[i].Time.Before(spikes[j].Time)
	})
	return spikes
}
