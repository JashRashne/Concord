package raft

import (
	"testing"

	"github.com/jashrashne/concord/internal/command"
)

func makeTestLeader(
	t *testing.T,
) (*State, uint64) {
	t.Helper()

	state := New()

	term, err :=
		state.StartElection("node-1")
	if err != nil {
		t.Fatal(err)
	}

	if !state.BecomeLeader(
		term,
		"node-1",
	) {
		t.Fatal(
			"expected state to become leader",
		)
	}

	return state, term
}

func TestInitializeLeaderReplication(
	t *testing.T,
) {
	state, _ := makeTestLeader(t)

	_, err := state.AppendLeaderCommand(
		command.Command{
			Type:  command.TypeSet,
			Key:   "a",
			Value: "1",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = state.AppendLeaderCommand(
		command.Command{
			Type:  command.TypeSet,
			Key:   "b",
			Value: "2",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	ok := state.InitializeLeaderReplication(
		[]string{
			"node-2",
			"node-3",
		},
	)

	if !ok {
		t.Fatal(
			"expected replication initialization",
		)
	}

	next, match, ok :=
		state.ReplicationProgress(
			"node-2",
		)

	if !ok {
		t.Fatal(
			"expected node-2 replication state",
		)
	}

	if next != 3 {
		t.Fatalf(
			"expected nextIndex 3, got %d",
			next,
		)
	}

	if match != 0 {
		t.Fatalf(
			"expected matchIndex 0, got %d",
			match,
		)
	}
}

func TestBackoffNextIndex(t *testing.T) {
	state, term := makeTestLeader(t)

	for _, key := range []string{"a", "b"} {
		_, err := state.AppendLeaderCommand(
			command.Command{
				Type:  command.TypeSet,
				Key:   key,
				Value: "value",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	state.InitializeLeaderReplication(
		[]string{"node-2"},
	)

	next, _, _ :=
		state.ReplicationProgress(
			"node-2",
		)

	if next != 3 {
		t.Fatalf(
			"expected nextIndex 3, got %d",
			next,
		)
	}

	if !state.BackoffNextIndex(
		"node-2",
		term,
		3,
	) {
		t.Fatal(
			"expected nextIndex to back off",
		)
	}

	next, _, _ =
		state.ReplicationProgress(
			"node-2",
		)

	if next != 2 {
		t.Fatalf(
			"expected nextIndex 2, got %d",
			next,
		)
	}
}

func TestRecordReplicationUpdatesProgress(
	t *testing.T,
) {
	state, term := makeTestLeader(t)

	for _, key := range []string{"a", "b"} {
		_, err := state.AppendLeaderCommand(
			command.Command{
				Type:  command.TypeSet,
				Key:   key,
				Value: "value",
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	state.InitializeLeaderReplication(
		[]string{"node-2"},
	)

	if !state.RecordReplication(
		"node-2",
		term,
		2,
	) {
		t.Fatal(
			"expected replication update",
		)
	}

	next, match, _ :=
		state.ReplicationProgress(
			"node-2",
		)

	if match != 2 {
		t.Fatalf(
			"expected matchIndex 2, got %d",
			match,
		)
	}

	if next != 3 {
		t.Fatalf(
			"expected nextIndex 3, got %d",
			next,
		)
	}
}

func TestCommitRequiresCurrentTermEntry(
	t *testing.T,
) {
	state := New()

	_, success, _, err :=
		state.HandleAppendEntries(
			1,
			"old-leader",
			0,
			0,
			[]LogEntry{
				testSetEntry(
					1,
					1,
					"old",
					"value",
				),
			},
		)

	if err != nil || !success {
		t.Fatal(
			"failed to prepare old-term entry",
		)
	}

	term, err :=
		state.StartElection("node-1")
	if err != nil {
		t.Fatal(err)
	}

	if !state.BecomeLeader(
		term,
		"node-1",
	) {
		t.Fatal(
			"expected node to become leader",
		)
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
		1,
	)

	commit :=
		state.AdvanceCommitIndex(3)

	if commit != 0 {
		t.Fatalf(
			"expected old-term entry not to commit directly, got %d",
			commit,
		)
	}

	entry, err :=
		state.AppendLeaderCommand(
			command.Command{
				Type:  command.TypeSet,
				Key:   "new",
				Value: "value",
			},
		)
	if err != nil {
		t.Fatal(err)
	}

	state.RecordReplication(
		"node-2",
		term,
		entry.Index,
	)

	commit =
		state.AdvanceCommitIndex(3)

	if commit != entry.Index {
		t.Fatalf(
			"expected commit index %d, got %d",
			entry.Index,
			commit,
		)
	}
}
