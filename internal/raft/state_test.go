package raft

import (
	"path/filepath"
	"testing"
)

func TestNewStateStartsAsFollower(t *testing.T) {
	state := New()

	got := state.Snapshot()

	if got.Role != RoleFollower {
		t.Fatalf(
			"expected FOLLOWER, got %s",
			got.Role,
		)
	}

	if got.CurrentTerm != 0 {
		t.Fatalf(
			"expected term 0, got %d",
			got.CurrentTerm,
		)
	}

	if got.VotedFor != "" {
		t.Fatalf(
			"expected no vote, got %s",
			got.VotedFor,
		)
	}
}

func TestRequestVoteRejectsStaleTerm(t *testing.T) {
	state := New()

	_, granted, err := state.HandleRequestVote(
		5,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !granted {
		t.Fatal("expected vote in term 5")
	}

	term, granted, err := state.HandleRequestVote(
		4,
		"node-3",
	)
	if err != nil {
		t.Fatal(err)
	}

	if granted {
		t.Fatal("expected stale vote request to be rejected")
	}

	if term != 5 {
		t.Fatalf(
			"expected current term 5, got %d",
			term,
		)
	}
}

func TestStateVotesOncePerTerm(t *testing.T) {
	state := New()

	_, granted, err := state.HandleRequestVote(
		3,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !granted {
		t.Fatal("expected first vote to be granted")
	}

	_, granted, err = state.HandleRequestVote(
		3,
		"node-3",
	)
	if err != nil {
		t.Fatal(err)
	}

	if granted {
		t.Fatal(
			"expected second candidate in same term to be rejected",
		)
	}
}

func TestStateAllowsRepeatedVoteForSameCandidate(
	t *testing.T,
) {
	state := New()

	_, granted, err := state.HandleRequestVote(
		3,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !granted {
		t.Fatal("expected first vote")
	}

	_, granted, err = state.HandleRequestVote(
		3,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !granted {
		t.Fatal(
			"expected repeated vote request from same candidate to succeed",
		)
	}
}

func TestHigherTermRequestForcesFollower(t *testing.T) {
	state := New()

	state.mu.Lock()
	state.role = RoleLeader
	state.currentTerm = 4
	state.mu.Unlock()

	term, granted, err := state.HandleRequestVote(
		5,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !granted {
		t.Fatal("expected vote to be granted")
	}

	if term != 5 {
		t.Fatalf(
			"expected term 5, got %d",
			term,
		)
	}

	got := state.Snapshot()

	if got.Role != RoleFollower {
		t.Fatalf(
			"expected FOLLOWER, got %s",
			got.Role,
		)
	}
}

func TestPersistentVoteSurvivesRestart(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"raft-state.json",
	)

	state1, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	_, granted, err := state1.HandleRequestVote(
		7,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !granted {
		t.Fatal("expected vote to be granted")
	}

	state2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}

	got := state2.Snapshot()

	if got.CurrentTerm != 7 {
		t.Fatalf(
			"expected recovered term 7, got %d",
			got.CurrentTerm,
		)
	}

	if got.VotedFor != "node-2" {
		t.Fatalf(
			"expected recovered vote for node-2, got %s",
			got.VotedFor,
		)
	}

	if got.Role != RoleFollower {
		t.Fatalf(
			"expected restarted node to be FOLLOWER, got %s",
			got.Role,
		)
	}

	_, granted, err = state2.HandleRequestVote(
		7,
		"node-3",
	)
	if err != nil {
		t.Fatal(err)
	}

	if granted {
		t.Fatal(
			"restarted node must not vote twice in term 7",
		)
	}
}

func TestStartElection(t *testing.T) {
	state := New()

	term, err := state.StartElection("node-1")
	if err != nil {
		t.Fatal(err)
	}

	if term != 1 {
		t.Fatalf(
			"expected term 1, got %d",
			term,
		)
	}

	got := state.Snapshot()

	if got.Role != RoleCandidate {
		t.Fatalf(
			"expected CANDIDATE, got %s",
			got.Role,
		)
	}

	if got.VotedFor != "node-1" {
		t.Fatalf(
			"expected self vote for node-1, got %s",
			got.VotedFor,
		)
	}
}

func TestCandidateBecomesLeader(t *testing.T) {
	state := New()

	term, err := state.StartElection("node-1")
	if err != nil {
		t.Fatal(err)
	}

	if !state.BecomeLeader(term, "node-1") {
		t.Fatal("expected candidate to become leader")
	}

	got := state.Snapshot()

	if got.Role != RoleLeader {
		t.Fatalf(
			"expected LEADER, got %s",
			got.Role,
		)
	}
}

func TestOldElectionCannotBecomeLeader(t *testing.T) {
	state := New()

	oldTerm, err := state.StartElection("node-1")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := state.ObserveTerm(oldTerm + 1); err != nil {
		t.Fatal(err)
	}

	if state.BecomeLeader(oldTerm, "node-1") {
		t.Fatal(
			"expected stale election to be unable to become leader",
		)
	}
}

func TestObserveHigherTermStepsDown(t *testing.T) {
	state := New()

	term, err := state.StartElection("node-1")
	if err != nil {
		t.Fatal(err)
	}

	if !state.BecomeLeader(term, "node-1") {
		t.Fatal("expected leader transition")
	}

	changed, err := state.ObserveTerm(term + 1)
	if err != nil {
		t.Fatal(err)
	}

	if !changed {
		t.Fatal("expected higher term to be observed")
	}

	got := state.Snapshot()

	if got.Role != RoleFollower {
		t.Fatalf(
			"expected FOLLOWER, got %s",
			got.Role,
		)
	}

	if got.VotedFor != "" {
		t.Fatalf(
			"expected vote to be cleared, got %s",
			got.VotedFor,
		)
	}
}

func TestStaleAppendEntriesRejected(t *testing.T) {
	state := New()

	if _, err := state.ObserveTerm(5); err != nil {
		t.Fatal(err)
	}

	term, success, err := state.HandleAppendEntries(
		4,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	if success {
		t.Fatal("expected stale heartbeat to be rejected")
	}

	if term != 5 {
		t.Fatalf(
			"expected term 5, got %d",
			term,
		)
	}
}
func TestCandidateStepsDownForCurrentTermLeader(
	t *testing.T,
) {
	state := New()

	term, err := state.StartElection("node-1")
	if err != nil {
		t.Fatal(err)
	}

	_, success, err := state.HandleAppendEntries(
		term,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !success {
		t.Fatal("expected heartbeat to succeed")
	}

	got := state.Snapshot()

	if got.Role != RoleFollower {
		t.Fatalf(
			"expected FOLLOWER, got %s",
			got.Role,
		)
	}

	if got.LeaderID != "node-2" {
		t.Fatalf(
			"expected leader node-2, got %s",
			got.LeaderID,
		)
	}
}

func TestHigherTermHeartbeatUpdatesState(t *testing.T) {
	state := New()

	_, _, err := state.HandleRequestVote(
		4,
		"node-2",
	)
	if err != nil {
		t.Fatal(err)
	}

	_, success, err := state.HandleAppendEntries(
		5,
		"node-3",
	)
	if err != nil {
		t.Fatal(err)
	}

	if !success {
		t.Fatal("expected heartbeat to succeed")
	}

	got := state.Snapshot()

	if got.CurrentTerm != 5 {
		t.Fatalf(
			"expected term 5, got %d",
			got.CurrentTerm,
		)
	}

	if got.Role != RoleFollower {
		t.Fatalf(
			"expected FOLLOWER, got %s",
			got.Role,
		)
	}

	if got.VotedFor != "" {
		t.Fatalf(
			"expected vote to be cleared, got %s",
			got.VotedFor,
		)
	}

	if got.LeaderID != "node-3" {
		t.Fatalf(
			"expected node-3 as leader, got %s",
			got.LeaderID,
		)
	}
}
