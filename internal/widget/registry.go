package widget

// Registry holds registered widgets by name.
type Registry struct {
	widgets map[string]Widget
}

// NewRegistry creates an empty widget registry.
func NewRegistry() *Registry {
	return &Registry{widgets: make(map[string]Widget)}
}

// Register adds a widget to the registry, keyed by its Name().
func (r *Registry) Register(w Widget) {
	r.widgets[w.Name()] = w
}

// Get retrieves a widget by name. Returns false if not found.
func (r *Registry) Get(name string) (Widget, bool) {
	w, ok := r.widgets[name]
	return w, ok
}

// Names returns the names of all registered widgets.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.widgets))
	for name := range r.widgets {
		names = append(names, name)
	}
	return names
}
