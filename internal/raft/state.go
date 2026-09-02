package raft

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Role string

const (
	RoleFollower  Role = "FOLLOWER"
	RoleCandidate Role = "CANDIDATE"
	RoleLeader    Role = "LEADER"
)

type Snapshot struct {
	Role        Role
	CurrentTerm uint64
	VotedFor    string
	LeaderID    string
}

type persistentState struct {
	CurrentTerm uint64 `json:"current_term"`
	VotedFor    string `json:"voted_for,omitempty"`
}

type State struct {
	mu sync.Mutex

	role        Role
	currentTerm uint64
	votedFor    string

	path     string
	leaderID string
}

func New() *State {
	return &State{
		role: RoleFollower,
	}
}

func Open(path string) (*State, error) {
	if path == "" {
		return nil, errors.New("Raft state path cannot be empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf(
			"create Raft state directory: %w",
			err,
		)
	}

	state := New()
	state.path = path

	data, err := os.ReadFile(path)

	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}

	if err != nil {
		return nil, fmt.Errorf(
			"read Raft state: %w",
			err,
		)
	}

	var stored persistentState

	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, fmt.Errorf(
			"decode Raft state: %w",
			err,
		)
	}

	state.currentTerm = stored.CurrentTerm
	state.votedFor = stored.VotedFor

	return state, nil
}

func (s *State) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	return Snapshot{
		Role:        s.role,
		CurrentTerm: s.currentTerm,
		VotedFor:    s.votedFor,
		LeaderID:    s.leaderID,
	}
}

func (s *State) HandleRequestVote(
	term uint64,
	candidateID string,
) (uint64, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if candidateID == "" {
		return s.currentTerm, false,
			errors.New("candidate ID cannot be empty")
	}

	if term < s.currentTerm {
		return s.currentTerm, false, nil
	}

	if term > s.currentTerm {
		oldTerm := s.currentTerm
		oldVotedFor := s.votedFor
		oldRole := s.role
		oldLeaderID := s.leaderID

		s.currentTerm = term
		s.votedFor = ""
		s.role = RoleFollower
		s.leaderID = ""

		if err := s.persistLocked(); err != nil {
			s.currentTerm = oldTerm
			s.votedFor = oldVotedFor
			s.role = oldRole
			s.leaderID = oldLeaderID

			return oldTerm, false, err
		}
	}

	if s.votedFor != "" && s.votedFor != candidateID {
		return s.currentTerm, false, nil
	}

	if s.votedFor == candidateID {
		return s.currentTerm, true, nil
	}

	oldVotedFor := s.votedFor

	s.votedFor = candidateID

	if err := s.persistLocked(); err != nil {
		s.votedFor = oldVotedFor
		return s.currentTerm, false, err
	}

	return s.currentTerm, true, nil
}

func (s *State) persistLocked() error {
	if s.path == "" {
		return nil
	}

	stored := persistentState{
		CurrentTerm: s.currentTerm,
		VotedFor:    s.votedFor,
	}

	data, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf(
			"encode Raft state: %w",
			err,
		)
	}

	data = append(data, '\n')

	tempPath := s.path + ".tmp"

	file, err := os.OpenFile(
		tempPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return fmt.Errorf(
			"open temporary Raft state: %w",
			err,
		)
	}

	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(tempPath)

		return fmt.Errorf(
			"write Raft state: %w",
			err,
		)
	}

	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tempPath)

		return fmt.Errorf(
			"sync Raft state: %w",
			err,
		)
	}

	if err := file.Close(); err != nil {
		os.Remove(tempPath)

		return fmt.Errorf(
			"close Raft state: %w",
			err,
		)
	}

	if err := os.Rename(tempPath, s.path); err != nil {
		os.Remove(tempPath)

		return fmt.Errorf(
			"replace Raft state: %w",
			err,
		)
	}

	dir, err := os.Open(filepath.Dir(s.path))
	if err != nil {
		return fmt.Errorf(
			"open Raft state directory: %w",
			err,
		)
	}
	defer dir.Close()

	if err := dir.Sync(); err != nil {
		return fmt.Errorf(
			"sync Raft state directory: %w",
			err,
		)
	}

	return nil
}

func (s *State) StartElection(
	candidateID string,
) (uint64, error) {
	if candidateID == "" {
		return 0, errors.New("candidate ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	oldRole := s.role
	oldTerm := s.currentTerm
	oldVotedFor := s.votedFor
	oldLeaderID := s.leaderID

	s.role = RoleCandidate
	s.currentTerm++
	s.votedFor = candidateID
	s.leaderID = ""

	if err := s.persistLocked(); err != nil {
		s.role = oldRole
		s.currentTerm = oldTerm
		s.votedFor = oldVotedFor
		s.leaderID = oldLeaderID

		return 0, err
	}

	return s.currentTerm, nil
}

func (s *State) BecomeLeader(
	term uint64,
	leaderID string,
) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if leaderID == "" {
		return false
	}

	if s.role != RoleCandidate {
		return false
	}

	if s.currentTerm != term {
		return false
	}

	s.role = RoleLeader
	s.leaderID = leaderID

	return true
}

func (s *State) ObserveTerm(term uint64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if term <= s.currentTerm {
		return false, nil
	}

	oldRole := s.role
	oldTerm := s.currentTerm
	oldVotedFor := s.votedFor
	oldLeaderID := s.leaderID

	s.currentTerm = term
	s.votedFor = ""
	s.role = RoleFollower
	s.leaderID = ""

	if err := s.persistLocked(); err != nil {
		s.role = oldRole
		s.currentTerm = oldTerm
		s.votedFor = oldVotedFor
		s.leaderID = oldLeaderID

		return false, err
	}

	return true, nil
}

func (s *State) HandleAppendEntries(
	term uint64,
	leaderID string,
) (uint64, bool, error) {
	if leaderID == "" {
		return 0, false,
			errors.New("leader ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if term < s.currentTerm {
		return s.currentTerm, false, nil
	}

	if term > s.currentTerm {
		oldRole := s.role
		oldTerm := s.currentTerm
		oldVotedFor := s.votedFor
		oldLeaderID := s.leaderID

		s.currentTerm = term
		s.votedFor = ""
		s.role = RoleFollower
		s.leaderID = leaderID

		if err := s.persistLocked(); err != nil {
			s.role = oldRole
			s.currentTerm = oldTerm
			s.votedFor = oldVotedFor
			s.leaderID = oldLeaderID

			return oldTerm, false, err
		}

		return s.currentTerm, true, nil
	}

	// Same term: a valid leader exists for this term.
	s.role = RoleFollower
	s.leaderID = leaderID

	return s.currentTerm, true, nil
}
