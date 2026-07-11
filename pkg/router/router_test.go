package router

import (
	"errors"
	"testing"
	"time"

	"github.com/DaviMGDev/unipm/pkg/adapter"
)

// mockAdapter is a test implementation of adapter.PackageManager that
// returns configurable responses.
type mockAdapter struct {
	name      string
	available bool
	searchFn  func(query string) ([]adapter.Package, error)
}

func (m *mockAdapter) Name() string                               { return m.name }
func (m *mockAdapter) IsAvailable() bool                          { return m.available }
func (m *mockAdapter) Search(q string) ([]adapter.Package, error) { return m.searchFn(q) }
func (m *mockAdapter) Install(p adapter.Package) error            { return nil }
func (m *mockAdapter) Uninstall(p adapter.Package) error          { return nil }
func (m *mockAdapter) Info(p adapter.Package) (adapter.Details, error) {
	return adapter.Details{}, nil
}

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
	if !r.IsEmpty() {
		t.Error("new registry should be empty")
	}
	if r.Count() != 0 {
		t.Errorf("Count() = %d, want 0", r.Count())
	}
}

func TestRegister(t *testing.T) {
	r := New()
	a := &mockAdapter{name: "apt"}

	r.Register(a)

	if r.IsEmpty() {
		t.Error("registry should not be empty after Register()")
	}
	if r.Count() != 1 {
		t.Errorf("Count() = %d, want 1", r.Count())
	}
}

func TestRegister_ReplaceExisting(t *testing.T) {
	r := New()
	a1 := &mockAdapter{name: "apt"}
	a2 := &mockAdapter{name: "apt"}

	r.Register(a1)
	r.Register(a2)

	if r.Count() != 1 {
		t.Errorf("Count() = %d, want 1 (duplicate name replaces)", r.Count())
	}
}

func TestGet_Found(t *testing.T) {
	r := New()
	r.Register(&mockAdapter{name: "apt"})

	a, err := r.Get("apt")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}
	if a.Name() != "apt" {
		t.Errorf("Name() = %q, want %q", a.Name(), "apt")
	}
}

func TestGet_NotFound(t *testing.T) {
	r := New()
	r.Register(&mockAdapter{name: "apt"})

	_, err := r.Get("npm")
	if err == nil {
		t.Error("Get() should return error for unregistered name")
	}
}

func TestNames_Sorted(t *testing.T) {
	r := New()
	r.Register(&mockAdapter{name: "npm"})
	r.Register(&mockAdapter{name: "apt"})
	r.Register(&mockAdapter{name: "brew"})

	names := r.Names()
	if len(names) != 3 {
		t.Fatalf("Names() length = %d, want 3", len(names))
	}

	// Must be sorted alphabetically (ADR-0002)
	expected := []string{"apt", "brew", "npm"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("Names()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestList_Alphabetical(t *testing.T) {
	r := New()
	r.Register(&mockAdapter{name: "npm"})
	r.Register(&mockAdapter{name: "apt"})

	adapters := r.List()
	if len(adapters) != 2 {
		t.Fatalf("List() length = %d, want 2", len(adapters))
	}
	if adapters[0].Name() != "apt" {
		t.Errorf("List()[0].Name() = %q, want %q (must be alphabetical)", adapters[0].Name(), "apt")
	}
	if adapters[1].Name() != "npm" {
		t.Errorf("List()[1].Name() = %q, want %q (must be alphabetical)", adapters[1].Name(), "npm")
	}
}

func TestIsEmpty(t *testing.T) {
	r := New()
	if !r.IsEmpty() {
		t.Error("IsEmpty() = false, want true for new registry")
	}

	r.Register(&mockAdapter{name: "apt"})
	if r.IsEmpty() {
		t.Error("IsEmpty() = true, want false after register")
	}
}

func TestCount(t *testing.T) {
	r := New()
	if r.Count() != 0 {
		t.Errorf("Count() = %d, want 0", r.Count())
	}

	r.Register(&mockAdapter{name: "apt"})
	if r.Count() != 1 {
		t.Errorf("Count() = %d, want 1", r.Count())
	}

	r.Register(&mockAdapter{name: "npm"})
	if r.Count() != 2 {
		t.Errorf("Count() = %d, want 2", r.Count())
	}
}

func TestSearchAll_Success(t *testing.T) {
	r := New()
	r.Register(&mockAdapter{
		name: "apt",
		searchFn: func(q string) ([]adapter.Package, error) {
			return []adapter.Package{
				{Name: "htop", Source: "apt", Version: "3.4.1", Description: "process viewer"},
			}, nil
		},
	})
	r.Register(&mockAdapter{
		name: "npm",
		searchFn: func(q string) ([]adapter.Package, error) {
			return []adapter.Package{
				{Name: "htop", Source: "npm", Version: "1.0.1", Description: "handle-to-promise"},
			}, nil
		},
	})

	pkgs, timedOut, errored := r.SearchAll("htop", 5*time.Second)

	if len(timedOut) != 0 {
		t.Errorf("timedOut = %v, want empty", timedOut)
	}
	if len(errored) != 0 {
		t.Errorf("errored = %v, want empty", errored)
	}
	if len(pkgs) != 2 {
		t.Fatalf("len(pkgs) = %d, want 2", len(pkgs))
	}

	// Results should be deduplicated
	sources := make(map[string]bool)
	for _, p := range pkgs {
		sources[p.Source] = true
	}
	if len(sources) != 2 {
		t.Errorf("expected results from 2 sources, got %d sources", len(sources))
	}
}

func TestSearchAll_Timeout(t *testing.T) {
	r := New()
	r.Register(&mockAdapter{
		name: "slow-adapter",
		searchFn: func(q string) ([]adapter.Package, error) {
			time.Sleep(2 * time.Second)
			return nil, nil
		},
	})

	pkgs, timedOut, _ := r.SearchAll("test", 10*time.Millisecond)

	if len(timedOut) != 1 {
		t.Errorf("timedOut length = %d, want 1", len(timedOut))
	}
	if timedOut[0] != "slow-adapter" {
		t.Errorf("timedOut[0] = %q, want %q", timedOut[0], "slow-adapter")
	}
	if len(pkgs) != 0 {
		t.Errorf("expected no results on timeout, got %d", len(pkgs))
	}
}

func TestSearchAll_Error(t *testing.T) {
	r := New()
	r.Register(&mockAdapter{
		name: "broken",
		searchFn: func(q string) ([]adapter.Package, error) {
			return nil, errors.New("connection refused")
		},
	})

	pkgs, timedOut, errored := r.SearchAll("test", 5*time.Second)

	if len(errored) != 1 {
		t.Errorf("errored length = %d, want 1", len(errored))
	}
	if errored[0] != "broken" {
		t.Errorf("errored[0] = %q, want %q", errored[0], "broken")
	}
	if len(timedOut) != 0 {
		t.Errorf("timedOut = %v, want empty", timedOut)
	}
	if len(pkgs) != 0 {
		t.Errorf("expected no results on error, got %d", len(pkgs))
	}
}

func TestSearchAll_EmptyRegistry(t *testing.T) {
	r := New()

	pkgs, timedOut, errored := r.SearchAll("test", 5*time.Second)

	if pkgs != nil {
		t.Error("expected nil packages for empty registry")
	}
	if timedOut != nil {
		t.Error("expected nil timedOut for empty registry")
	}
	if errored != nil {
		t.Error("expected nil errored for empty registry")
	}
}

func TestSearchAll_Deduplication(t *testing.T) {
	r := New()
	r.Register(&mockAdapter{
		name: "test-adapter",
		searchFn: func(q string) ([]adapter.Package, error) {
			return []adapter.Package{
				{Name: "dupe", Source: "test-adapter", Version: "1.0"},
				{Name: "dupe", Source: "test-adapter", Version: "1.0"},
				{Name: "unique", Source: "test-adapter", Version: "2.0"},
			}, nil
		},
	})

	pkgs, _, _ := r.SearchAll("test", 5*time.Second)

	if len(pkgs) != 2 {
		t.Fatalf("len(pkgs) = %d, want 2 (deduplicated)", len(pkgs))
	}

	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}

	hasDupe := false
	hasUnique := false
	for _, n := range names {
		if n == "dupe" {
			hasDupe = true
		}
		if n == "unique" {
			hasUnique = true
		}
	}
	if !hasDupe || !hasUnique {
		t.Errorf("expected both 'dupe' and 'unique', got %v", names)
	}
}

func TestDeduplicate(t *testing.T) {
	pkgs := []adapter.Package{
		{Name: "a", Source: "apt"},
		{Name: "a", Source: "apt"}, // duplicate
		{Name: "a", Source: "npm"}, // same name, different source = not duplicate
		{Name: "b", Source: "apt"},
	}

	result := deduplicate(pkgs)

	if len(result) != 3 {
		t.Fatalf("len(result) = %d, want 3", len(result))
	}
}

func TestIsTimeoutError(t *testing.T) {
	if !isTimeoutError(errors.New("timeout after 10s")) {
		t.Error("isTimeoutError() = false for timeout error")
	}
	if isTimeoutError(errors.New("connection refused")) {
		t.Error("isTimeoutError() = true for non-timeout error")
	}
	if isTimeoutError(nil) {
		t.Error("isTimeoutError() = true for nil")
	}
}
