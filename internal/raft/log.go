package raft

import (
	"fmt"

	"github.com/jashrashne/concord/internal/command"
)

type LogEntry struct {
	Index   uint64          `json:"index"`
	Term    uint64          `json:"term"`
	Command command.Command `json:"command"`
}

func (s *State) LastLogInfo() (uint64, uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.lastLogInfoLocked()
}

func (s *State) lastLogInfoLocked() (uint64, uint64) {
	if len(s.log) == 0 {
		return 0, 0
	}

	last := s.log[len(s.log)-1]

	return last.Index, last.Term
}

func (s *State) LogEntries() []LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := make([]LogEntry, len(s.log))
	copy(entries, s.log)

	return entries
}

func (s *State) appendEntriesLocked(
	prevLogIndex uint64,
	prevLogTerm uint64,
	entries []LogEntry,
) (uint64, bool, error) {
	if prevLogIndex > uint64(len(s.log)) {
		return 0, false, nil
	}

	if prevLogIndex > 0 {
		previous := s.log[prevLogIndex-1]

		if previous.Term != prevLogTerm {
			return 0, false, nil
		}
	}

	for i, entry := range entries {
		expectedIndex :=
			prevLogIndex + uint64(i) + 1

		if entry.Index != expectedIndex {
			return 0, false, fmt.Errorf(
				"expected log entry index %d, got %d",
				expectedIndex,
				entry.Index,
			)
		}

		if entry.Term == 0 {
			return 0, false,
				fmt.Errorf(
					"log entry term cannot be zero",
				)
		}

		if err := entry.Command.Validate(); err != nil {
			return 0, false,
				fmt.Errorf(
					"invalid log command: %w",
					err,
				)
		}
	}

	for i, entry := range entries {
		position := entry.Index - 1

		if position < uint64(len(s.log)) {
			local := s.log[position]

			if local.Term == entry.Term {
				continue
			}

			s.log = s.log[:position]
			s.log = append(s.log, entries[i:]...)

			return entry.Index +
					uint64(len(entries[i:])) - 1,
				true,
				nil
		}

		s.log = append(s.log, entries[i:]...)

		break
	}

	return prevLogIndex + uint64(len(entries)),
		true,
		nil
}
