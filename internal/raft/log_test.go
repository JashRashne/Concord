package raft

import (
	"testing"

	"github.com/jashrashne/concord/internal/command"
)

func testSetEntry(
	index uint64,
	term uint64,
	key string,
	value string,
) LogEntry {
	return LogEntry{
		Index: index,
		Term:  term,
		Command: command.Command{
			Type:  command.TypeSet,
			Key:   key,
			Value: value,
		},
	}
}

func TestAppendEntriesToEmptyLog(t *testing.T) {
	state := New()

	term, success, matchIndex, err :=
		state.HandleAppendEntries(
			1,
			"node-1",
			0,
			0,
			[]LogEntry{
				testSetEntry(1, 1, "x", "10"),
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if !success {
		t.Fatal("expected AppendEntries success")
	}

	if term != 1 {
		t.Fatalf("expected term 1, got %d", term)
	}

	if matchIndex != 1 {
		t.Fatalf(
			"expected match index 1, got %d",
			matchIndex,
		)
	}

	entries := state.LogEntries()

	if len(entries) != 1 {
		t.Fatalf(
			"expected 1 entry, got %d",
			len(entries),
		)
	}
}

func TestAppendEntriesRejectsMissingPrevIndex(
	t *testing.T,
) {
	state := New()

	_, success, _, err :=
		state.HandleAppendEntries(
			1,
			"node-1",
			3,
			1,
			nil,
		)

	if err != nil {
		t.Fatal(err)
	}

	if success {
		t.Fatal(
			"expected missing previous index to be rejected",
		)
	}
}

func TestAppendEntriesRejectsPrevTermMismatch(
	t *testing.T,
) {
	state := New()

	_, success, _, err :=
		state.HandleAppendEntries(
			1,
			"node-1",
			0,
			0,
			[]LogEntry{
				testSetEntry(1, 1, "x", "10"),
			},
		)
	if err != nil || !success {
		t.Fatal("failed to prepare log")
	}

	_, success, _, err =
		state.HandleAppendEntries(
			2,
			"node-2",
			1,
			99,
			nil,
		)

	if err != nil {
		t.Fatal(err)
	}

	if success {
		t.Fatal(
			"expected previous term mismatch to fail",
		)
	}
}
func TestAppendEntriesReplacesConflictingSuffix(
	t *testing.T,
) {
	state := New()

	_, success, _, err :=
		state.HandleAppendEntries(
			2,
			"node-1",
			0,
			0,
			[]LogEntry{
				testSetEntry(1, 1, "a", "1"),
				testSetEntry(2, 1, "b", "2"),
				testSetEntry(3, 2, "old", "value"),
			},
		)

	if err != nil || !success {
		t.Fatal("failed to prepare follower log")
	}

	_, success, _, err =
		state.HandleAppendEntries(
			3,
			"node-2",
			2,
			1,
			[]LogEntry{
				testSetEntry(3, 3, "new", "value"),
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	if !success {
		t.Fatal("expected replacement to succeed")
	}

	entries := state.LogEntries()

	if len(entries) != 3 {
		t.Fatalf(
			"expected 3 entries, got %d",
			len(entries),
		)
	}

	if entries[2].Term != 3 {
		t.Fatalf(
			"expected entry 3 term 3, got %d",
			entries[2].Term,
		)
	}

	if entries[2].Command.Key != "new" {
		t.Fatalf(
			"expected replacement command, got key %s",
			entries[2].Command.Key,
		)
	}
}
