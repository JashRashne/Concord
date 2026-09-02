package raft

import "testing"

func TestMajority(t *testing.T) {
	tests := []struct {
		clusterSize int
		expected    int
	}{
		{1, 1},
		{3, 2},
		{5, 3},
		{7, 4},
	}

	for _, test := range tests {
		got := Majority(test.clusterSize)

		if got != test.expected {
			t.Fatalf(
				"cluster size %d: expected majority %d, got %d",
				test.clusterSize,
				test.expected,
				got,
			)
		}
	}
}
