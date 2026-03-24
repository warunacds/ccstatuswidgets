package widgets

import "github.com/warunacds/ccstatuswidgets/internal/widget"

// RegisterAll registers all built-in widgets with the given registry.
func RegisterAll(r *widget.Registry) {
	// Phase 1 widgets.
	r.Register(&ModelWidget{})
	r.Register(&EffortWidget{})
	r.Register(&DirectoryWidget{})
	r.Register(&GitBranchWidget{})
	r.Register(&GitStatusWidget{})
	r.Register(&ContextBarWidget{})
	r.Register(&Usage5hWidget{})
	r.Register(&Usage7dWidget{})
	r.Register(&LinesWidget{})
	r.Register(&CostWidget{})
	r.Register(&MemoryWidget{})

	// Phase 2 widgets.
	r.Register(&MoonWidget{})
	r.Register(&NowPlayingWidget{})
	r.Register(&WeatherWidget{})
	r.Register(&FlightWidget{})
	r.Register(&CricketWidget{})
	r.Register(&StocksWidget{})
	r.Register(&HackernewsWidget{})
	r.Register(&PomodoroWidget{})
	r.Register(&SessionTimeWidget{})
}
