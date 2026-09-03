package raft

import (
	"testing"

	"github.com/jashrashne/concord/internal/command"
)

func TestCommittedEntryBecomesAvailableToApply(
	t *testing.T,
) {
	state, term := makeTestLeader(t)

	entry, err :=
		state.AppendLeaderCommand(
			command.Command{
				Type:  command.TypeSet,
				Key:   "name",
				Value: "alice",
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	state.InitializeLeaderReplication(
		[]string{
			"node-2",
			"node-3",
		},
	)

	state.RecordReplication(
		"node-2",
		term,
		entry.Index,
	)

	state.AdvanceCommitIndex(3)

	next, ok :=
		state.NextCommittedEntry()

	if !ok {
		t.Fatal(
			"expected committed entry",
		)
	}

	if next.Index != entry.Index {
		t.Fatalf(
			"expected index %d, got %d",
			entry.Index,
			next.Index,
		)
	}

	if err :=
		state.MarkApplied(
			next.Index,
		); err != nil {
		t.Fatal(err)
	}

	_, ok =
		state.NextCommittedEntry()

	if ok {
		t.Fatal(
			"expected no additional committed entries",
		)
	}
}
