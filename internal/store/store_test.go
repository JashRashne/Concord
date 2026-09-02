package store

import (
	"fmt"
	"sync"
	"testing"
)

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

func TestConcurrentAccess(t *testing.T) {
	s := New()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)

			s.Set(key, value)
			s.Get(key)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		expected := fmt.Sprintf("value-%d", i)

		value, ok := s.Get(key)

		if !ok {
			t.Fatalf("expected %s to exist", key)
		}

		if value != expected {
			t.Fatalf("expected %s, got %s", expected, value)
		}
	}
}

func TestConcurrentDelete(t *testing.T) {
	s := New()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		s.Set(key, "value")
	}

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", i)

			s.Get(key)
			s.Delete(key)
		}(i)
	}

	wg.Wait()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)

		if _, ok := s.Get(key); ok {
			t.Fatalf("expected %s to be deleted", key)
		}
	}
}
