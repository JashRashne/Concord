package raft

import "testing"

func TestRandomElectionTimeoutWithinRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		timeout := RandomElectionTimeout()

		if timeout < MinElectionTimeout {
			t.Fatalf(
				"timeout %v below minimum %v",
				timeout,
				MinElectionTimeout,
			)
		}

		if timeout >= MaxElectionTimeout {
			t.Fatalf(
				"timeout %v above maximum %v",
				timeout,
				MaxElectionTimeout,
			)
		}
	}
}
