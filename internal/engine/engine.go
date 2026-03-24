// Package engine runs all configured widgets concurrently and collects their results.
package engine

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/cache"
	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
	"github.com/warunacds/ccstatuswidgets/internal/renderer"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
)

const cacheTTL = 5 * time.Minute

// Engine orchestrates concurrent widget execution with timeout and cache fallback.
type Engine struct {
	registry *widget.Registry
	cache    *cache.Cache
	timeout  time.Duration
}

// New creates an Engine with the given registry, cache, and default timeout.
func New(registry *widget.Registry, cache *cache.Cache, timeout time.Duration) *Engine {
	return &Engine{
		registry: registry,
		cache:    cache,
		timeout:  timeout,
	}
}

// widgetJob tracks where a widget result should be placed in the output.
type widgetJob struct {
	lineIdx   int
	slotIdx   int
	name      string
	w         widget.Widget
	widgetCfg map[string]interface{}
}

// Run executes all configured widgets concurrently and returns results grouped by line.
// Each widget has a timeout; on timeout or error, the last cached result is used as fallback.
// If no cached result exists, the output is nil. Order is preserved via indexed slots.
func (e *Engine) Run(input *protocol.StatusLineInput, cfg *config.Config) [][]renderer.WidgetResult {
	timeout := e.timeout
	if cfg.TimeoutMs > 0 {
		timeout = time.Duration(cfg.TimeoutMs) * time.Millisecond
	}

	// Build the result grid and collect jobs.
	results := make([][]renderer.WidgetResult, len(cfg.Lines))
	var jobs []widgetJob

	for lineIdx, line := range cfg.Lines {
		// Pre-filter: only include widgets found in the registry.
		var lineResults []renderer.WidgetResult
		slotIdx := 0
		for _, wName := range line.Widgets {
			w, ok := e.registry.Get(wName)
			if !ok {
				continue
			}
			lineResults = append(lineResults, renderer.WidgetResult{Name: wName})
			jobs = append(jobs, widgetJob{
				lineIdx:   lineIdx,
				slotIdx:   slotIdx,
				name:      wName,
				w:         w,
				widgetCfg: cfg.Widgets[wName],
			})
			slotIdx++
		}
		results[lineIdx] = lineResults
	}

	// Run all jobs concurrently.
	var wg sync.WaitGroup
	wg.Add(len(jobs))

	for _, job := range jobs {
		go func(j widgetJob) {
			defer wg.Done()
			output := e.executeWidget(j.w, input, j.widgetCfg, timeout)
			results[j.lineIdx][j.slotIdx].Output = output
		}(job)
	}

	wg.Wait()
	return results
}

// executeWidget runs a single widget with a timeout. On success the result is cached.
// On timeout or error, the cached value is returned (or nil if no cache entry exists).
func (e *Engine) executeWidget(w widget.Widget, input *protocol.StatusLineInput, widgetCfg map[string]interface{}, timeout time.Duration) *protocol.WidgetOutput {
	type result struct {
		output *protocol.WidgetOutput
		err    error
	}

	ch := make(chan result, 1)
	go func() {
		out, err := w.Render(input, widgetCfg)
		ch <- result{output: out, err: err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return e.fromCache(w.Name())
		}
		e.toCache(w.Name(), res.output)
		return res.output

	case <-time.After(timeout):
		return e.fromCache(w.Name())
	}
}

// fromCache retrieves a widget's last output from the cache.
// Returns nil if the entry is missing or expired.
func (e *Engine) fromCache(name string) *protocol.WidgetOutput {
	data, ok := e.cache.Get(name)
	if !ok {
		return nil
	}
	var out protocol.WidgetOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return &out
}

// toCache stores a widget's output in the cache with the standard TTL.
func (e *Engine) toCache(name string, out *protocol.WidgetOutput) {
	data, err := json.Marshal(out)
	if err != nil {
		return
	}
	e.cache.Set(name, data, cacheTTL)
}
