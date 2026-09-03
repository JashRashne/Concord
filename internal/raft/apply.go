package raft

import "fmt"

func (s *State) NextCommittedEntry() (LogEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lastApplied >= s.commitIndex {
		return LogEntry{}, false
	}

	next := s.lastApplied + 1

	if next > uint64(len(s.log)) {
		return LogEntry{}, false
	}

	return s.log[next-1], true
}

func (s *State) MarkApplied(index uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	expected := s.lastApplied + 1

	if index != expected {
		return fmt.Errorf(
			"expected applied index %d, got %d",
			expected,
			index,
		)
	}

	if index > s.commitIndex {
		return fmt.Errorf(
			"cannot apply uncommitted index %d",
			index,
		)
	}

	s.lastApplied = index

	return nil
}
