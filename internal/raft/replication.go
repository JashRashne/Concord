package raft

import (
	"fmt"

	"github.com/jashrashne/concord/internal/command"
)

type ReplicationBatch struct {
	Term         uint64
	NextIndex    uint64
	PrevLogIndex uint64
	PrevLogTerm  uint64
	Entries      []LogEntry
	LeaderCommit uint64
}

func (s *State) AppendLeaderCommand(
	cmd command.Command,
) (LogEntry, error) {
	if err := cmd.Validate(); err != nil {
		return LogEntry{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.role != RoleLeader {
		return LogEntry{},
			fmt.Errorf("only leader can append commands")
	}

	if s.currentTerm == 0 {
		return LogEntry{},
			fmt.Errorf("leader term cannot be zero")
	}

	entry := LogEntry{
		Index:   uint64(len(s.log)) + 1,
		Term:    s.currentTerm,
		Command: cmd,
	}

	s.log = append(s.log, entry)

	return entry, nil
}

func (s *State) InitializeLeaderReplication(
	peerIDs []string,
) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.role != RoleLeader {
		return false
	}

	lastIndex, _ := s.lastLogInfoLocked()

	s.nextIndex = make(
		map[string]uint64,
		len(peerIDs),
	)

	s.matchIndex = make(
		map[string]uint64,
		len(peerIDs),
	)

	for _, peerID := range peerIDs {
		s.nextIndex[peerID] = lastIndex + 1
		s.matchIndex[peerID] = 0
	}

	return true
}

func (s *State) ReplicationProgress(
	peerID string,
) (uint64, uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next, ok := s.nextIndex[peerID]
	if !ok {
		return 0, 0, false
	}

	return next, s.matchIndex[peerID], true
}

func (s *State) BuildReplicationBatch(
	peerID string,
	term uint64,
) (ReplicationBatch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.role != RoleLeader ||
		s.currentTerm != term {
		return ReplicationBatch{}, false
	}

	next, ok := s.nextIndex[peerID]
	if !ok {
		return ReplicationBatch{}, false
	}

	lastIndex := uint64(len(s.log))

	if next < 1 {
		next = 1
	}

	if next > lastIndex+1 {
		next = lastIndex + 1
		s.nextIndex[peerID] = next
	}

	prevLogIndex := next - 1

	var prevLogTerm uint64

	if prevLogIndex > 0 {
		prevLogTerm =
			s.log[prevLogIndex-1].Term
	}

	var entries []LogEntry

	if next <= lastIndex {
		entries = make(
			[]LogEntry,
			len(s.log)-int(next)+1,
		)

		copy(
			entries,
			s.log[next-1:],
		)
	}

	return ReplicationBatch{
		Term:         s.currentTerm,
		NextIndex:    next,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: s.commitIndex,
	}, true
}

func (s *State) RecordReplication(
	peerID string,
	term uint64,
	matchIndex uint64,
) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.role != RoleLeader ||
		s.currentTerm != term {
		return false
	}

	if _, ok := s.nextIndex[peerID]; !ok {
		return false
	}

	if matchIndex > uint64(len(s.log)) {
		return false
	}

	if matchIndex > s.matchIndex[peerID] {
		s.matchIndex[peerID] = matchIndex
	}

	next := s.matchIndex[peerID] + 1

	if next > s.nextIndex[peerID] {
		s.nextIndex[peerID] = next
	}

	return true
}

func (s *State) BackoffNextIndex(
	peerID string,
	term uint64,
	attemptedNext uint64,
) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.role != RoleLeader ||
		s.currentTerm != term {
		return false
	}

	current, ok := s.nextIndex[peerID]
	if !ok {
		return false
	}

	// Ignore an old failure if another RPC
	// already moved this follower forward.
	if current != attemptedNext {
		return true
	}

	if current <= 1 {
		return false
	}

	s.nextIndex[peerID] = current - 1

	return true
}

func (s *State) AdvanceCommitIndex(
	clusterSize int,
) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.role != RoleLeader ||
		clusterSize < 1 {
		return s.commitIndex
	}

	majority := clusterSize/2 + 1
	lastIndex := uint64(len(s.log))

	for index := lastIndex; index > s.commitIndex; index-- {

		// Raft's current-term commit rule.
		if s.log[index-1].Term != s.currentTerm {
			continue
		}

		replicas := 1 // Leader itself.

		for _, match := range s.matchIndex {
			if match >= index {
				replicas++
			}
		}

		if replicas >= majority {
			s.commitIndex = index
			break
		}
	}

	return s.commitIndex
}
