package wal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jashrashne/concord/internal/command"
)

func TestAppendAndReadAll(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"node-1",
		"wal.log",
	)

	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()

	setRecord := NewRecord(command.Command{
		Type:  command.TypeSet,
		Key:   "name",
		Value: "alice",
	})

	deleteRecord := NewRecord(command.Command{
		Type: command.TypeDelete,
		Key:  "name",
	})

	if err := log.Append(setRecord); err != nil {
		t.Fatal(err)
	}

	if err := log.Append(deleteRecord); err != nil {
		t.Fatal(err)
	}

	records, err := log.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 2 {
		t.Fatalf(
			"expected 2 records, got %d",
			len(records),
		)
	}

	if records[0].Command.Type != command.TypeSet {
		t.Fatalf(
			"expected first command to be SET, got %s",
			records[0].Command.Type,
		)
	}

	if records[0].Command.Value != "alice" {
		t.Fatalf(
			"expected alice, got %s",
			records[0].Command.Value,
		)
	}

	if records[1].Command.Type != command.TypeDelete {
		t.Fatalf(
			"expected second command to be DELETE, got %s",
			records[1].Command.Type,
		)
	}
}

func TestWALIsNewlineDelimited(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"wal.log",
	)

	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	record := NewRecord(command.Command{
		Type:  command.TypeSet,
		Key:   "name",
		Value: "alice",
	})

	if err := log.Append(record); err != nil {
		t.Fatal(err)
	}

	if err := log.Append(record); err != nil {
		t.Fatal(err)
	}

	if err := log.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := strings.Count(string(data), "\n"); got != 2 {
		t.Fatalf(
			"expected 2 newline-delimited records, got %d",
			got,
		)
	}
}
