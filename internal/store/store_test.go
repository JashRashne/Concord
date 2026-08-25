package store

import "testing"

func TestSetAndGet(t *testing.T) {
	s := New()

	s.Set("name", "alice")

	value, ok := s.Get("name")

	if !ok {
		t.Fatal("expected key to exist")
	}

	if value != "alice" {
		t.Fatalf("expected alice, got %s", value)
	}
}

func TestGetMissingKey(t *testing.T) {
	s := New()

	value, ok := s.Get("missing")

	if ok {
		t.Fatal("expected missing key to return false")
	}

	if value != "" {
		t.Fatalf("expected empty string, got %s", value)
	}
}

func TestDelete(t *testing.T) {
	s := New()

	s.Set("name", "alice")
	s.Delete("name")

	_, ok := s.Get("name")

	if ok {
		t.Fatal("expected deleted key to be missing")
	}
}
