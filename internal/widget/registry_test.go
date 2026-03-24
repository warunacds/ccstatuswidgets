package widget_test

import (
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/widget"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	r := widget.NewRegistry()
	r.Register(&mockWidget{})
	w, ok := r.Get("mock")
	if !ok {
		t.Fatal("widget not found")
	}
	if w.Name() != "mock" {
		t.Fatalf("expected mock, got %s", w.Name())
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := widget.NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestRegistryNames(t *testing.T) {
	r := widget.NewRegistry()
	r.Register(&mockWidget{})
	names := r.Names()
	if len(names) != 1 || names[0] != "mock" {
		t.Fatalf("unexpected names: %v", names)
	}
}
