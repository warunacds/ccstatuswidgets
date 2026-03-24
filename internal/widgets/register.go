package widgets

import "github.com/warunacds/ccstatuswidgets/internal/widget"

// RegisterAll registers all built-in widgets with the given registry.
func RegisterAll(r *widget.Registry) {
	r.Register(&ModelWidget{})
	r.Register(&EffortWidget{})
	r.Register(&DirectoryWidget{})
	r.Register(&GitBranchWidget{})
	r.Register(&ContextBarWidget{})
	r.Register(&Usage5hWidget{})
	r.Register(&Usage7dWidget{})
	r.Register(&LinesWidget{})
	r.Register(&CostWidget{})
	r.Register(&MemoryWidget{})
	r.Register(&MoonWidget{})
	r.Register(&NowPlayingWidget{})
}
