package raft

func Majority(clusterSize int) int {
	return clusterSize/2 + 1
}
