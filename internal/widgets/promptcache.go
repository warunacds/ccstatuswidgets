package widgets

import (
	"fmt"
	"math"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// PromptCacheWidget displays the prompt cache hit ratio and whether the cache is warm.
type PromptCacheWidget struct{}

func (w *PromptCacheWidget) Name() string {
	return "prompt-cache"
}

func (w *PromptCacheWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	pc := input.PromptCache
	if pc == nil || !pc.CachingObserved {
		return nil, nil
	}
	pct := int(math.Round(pc.HitRatio * 100))
	if pc.Warm {
		return &protocol.WidgetOutput{
			Text:  fmt.Sprintf("cache %d%%", pct),
			Color: "green",
		}, nil
	}
	return &protocol.WidgetOutput{
		Text:  fmt.Sprintf("cache %d%% cold", pct),
		Color: "dim",
	}, nil
}
