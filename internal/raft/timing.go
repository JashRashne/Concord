package raft

import (
	"math/rand"
	"time"
)

const (
	MinElectionTimeout = 800 * time.Millisecond
	MaxElectionTimeout = 1400 * time.Millisecond

	RequestVoteTimeout = 400 * time.Millisecond
)

func RandomElectionTimeout() time.Duration {
	delta := MaxElectionTimeout - MinElectionTimeout

	return MinElectionTimeout +
		time.Duration(rand.Int63n(int64(delta)))
}
