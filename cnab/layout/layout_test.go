package layout

import (
	"sync"
	"testing"
)

type stubLayout struct {
	name    string
	version string
	records map[RecordKey]RecordSpec
}

func (s stubLayout) Name() string    { return s.name }
func (s stubLayout) Version() string { return s.version }
func (s stubLayout) Record(key RecordKey) (RecordSpec, bool) {
	spec, ok := s.records[key]
	return spec, ok
}

func TestRegisterAndLookup(t *testing.T) {
	l := stubLayout{name: "test-register-lookup", version: "001"}
	Register(l.name, l)

	got, ok := Lookup(l.name)
	if !ok {
		t.Fatalf("Lookup(%q) returned ok=false, want true", l.name)
	}
	if got.Name() != l.name {
		t.Fatalf("Lookup(%q).Name() = %q, want %q", l.name, got.Name(), l.name)
	}
}

func TestLookupMissing(t *testing.T) {
	if _, ok := Lookup("does-not-exist"); ok {
		t.Fatalf("Lookup of an unregistered name returned ok=true")
	}
}

func TestRegisterNilPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Register(nil) did not panic")
		}
	}()
	Register("test-register-nil", nil)
}

func TestRegisterEmptyNamePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Register with empty name did not panic")
		}
	}()
	Register("", stubLayout{name: "x", version: "001"})
}

func TestRegisterEmptyVersionPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Register with empty version did not panic")
		}
	}()
	Register("test-register-empty-version", stubLayout{name: "x", version: ""})
}

func TestRegisterDuplicatePanics(t *testing.T) {
	l := stubLayout{name: "test-register-duplicate", version: "001"}
	Register(l.name, l)

	defer func() {
		if recover() == nil {
			t.Fatal("Register called twice for the same name did not panic")
		}
	}()
	Register(l.name, l)
}

func TestNamesIncludesRegistered(t *testing.T) {
	l := stubLayout{name: "test-names-includes", version: "001"}
	Register(l.name, l)

	found := false
	for _, n := range Names() {
		if n == l.name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Names() = %v, want it to contain %q", Names(), l.name)
	}
}

func TestConcurrentRegisterAndLookup(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := "test-concurrent-" + string(rune('a'+i))
			Register(name, stubLayout{name: name, version: "001"})
			if _, ok := Lookup(name); !ok {
				t.Errorf("Lookup(%q) after concurrent Register returned ok=false", name)
			}
		}()
	}
	wg.Wait()
}
