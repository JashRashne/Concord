package node

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/jashrashne/concord/internal/command"
	"github.com/jashrashne/concord/internal/peer"
	"github.com/jashrashne/concord/internal/protocol"
	"github.com/jashrashne/concord/internal/raft"
)

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	return port
}

func newThreeNodeTestCluster(t *testing.T) []*Node {
	t.Helper()

	ports := []int{
		freePort(t),
		freePort(t),
		freePort(t),
	}

	nodes := []*Node{
		New("node-1", ports[0]),
		New("node-2", ports[1]),
		New("node-3", ports[2]),
	}

	for i, n := range nodes {
		for j, other := range nodes {
			if i == j {
				continue
			}

			n.AddPeer(peer.Peer{
				ID: other.ID,
				Address: fmt.Sprintf(
					"127.0.0.1:%d",
					ports[j],
				),
			})
		}
	}

	return nodes
}

func startTestCluster(t *testing.T, nodes []*Node) {
	t.Helper()

	for _, n := range nodes {
		n := n

		go func() {
			if err := n.Start(); err != nil {
				select {
				case <-n.stopCh:
					return
				default:
					fmt.Printf(
						"test node %s stopped with error: %v\n",
						n.ID,
						err,
					)
				}
			}
		}()
	}

	t.Cleanup(func() {
		for _, n := range nodes {
			_ = n.Close()
		}
	})
}

func waitForLeader(
	t *testing.T,
	nodes []*Node,
	timeout time.Duration,
) (*Node, raft.Snapshot) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		var leader *Node
		var leaderSnapshot raft.Snapshot
		leaderCount := 0

		for _, n := range nodes {
			snapshot := n.RaftSnapshot()

			if snapshot.Role == raft.RoleLeader {
				leader = n
				leaderSnapshot = snapshot
				leaderCount++
			}
		}

		if leaderCount == 1 {
			return leader, leaderSnapshot
		}

		time.Sleep(25 * time.Millisecond)
	}

	t.Fatal("timed out waiting for exactly one leader")

	return nil, raft.Snapshot{}
}

func TestThreeNodeClusterElectsLeader(t *testing.T) {
	nodes := newThreeNodeTestCluster(t)

	startTestCluster(t, nodes)

	leader, snapshot := waitForLeader(
		t,
		nodes,
		5*time.Second,
	)

	if leader == nil {
		t.Fatal("expected a leader")
	}

	if snapshot.CurrentTerm == 0 {
		t.Fatal("expected election term greater than zero")
	}

	if snapshot.LeaderID != leader.ID {
		t.Fatalf(
			"expected leader ID %s, got %s",
			leader.ID,
			snapshot.LeaderID,
		)
	}
}

func TestRequestVoteTimesOut(t *testing.T) {
	listener, err := net.Listen(
		"tcp",
		"127.0.0.1:0",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		time.Sleep(time.Second)
	}()

	n := New("node-1", 0)

	timeout := 100 * time.Millisecond

	start := time.Now()

	_, err = n.RequestVote(
		listener.Addr().String(),
		protocol.RequestVoteRequest{
			Term:        1,
			CandidateID: "node-1",
		},
		timeout,
	)

	elapsed := time.Since(start)

	if err == nil {
		t.Fatal(
			"expected RequestVote to time out",
		)
	}

	if elapsed > 500*time.Millisecond {
		t.Fatalf(
			"RequestVote took too long to time out: %v",
			elapsed,
		)
	}
}

func TestClusterElectsLeaderWithOneUnavailablePeer(
	t *testing.T,
) {
	port1 := freePort(t)
	port2 := freePort(t)
	deadPort := freePort(t)

	node1 := New("node-1", port1)
	node2 := New("node-2", port2)

	node1.AddPeer(peer.Peer{
		ID:      "node-2",
		Address: fmt.Sprintf("127.0.0.1:%d", port2),
	})

	node1.AddPeer(peer.Peer{
		ID: "node-3",
		Address: fmt.Sprintf(
			"127.0.0.1:%d",
			deadPort,
		),
	})

	node2.AddPeer(peer.Peer{
		ID:      "node-1",
		Address: fmt.Sprintf("127.0.0.1:%d", port1),
	})

	node2.AddPeer(peer.Peer{
		ID: "node-3",
		Address: fmt.Sprintf(
			"127.0.0.1:%d",
			deadPort,
		),
	})

	nodes := []*Node{node1, node2}

	startTestCluster(t, nodes)

	leader, _ := waitForLeader(
		t,
		nodes,
		5*time.Second,
	)

	if leader == nil {
		t.Fatal(
			"expected two available nodes to elect a leader",
		)
	}
}

func TestLeaderRemainsStableWithHealthyCluster(
	t *testing.T,
) {
	nodes := newThreeNodeTestCluster(t)

	startTestCluster(t, nodes)

	leader, first := waitForLeader(
		t,
		nodes,
		5*time.Second,
	)

	time.Sleep(
		raft.MaxElectionTimeout +
			2*raft.HeartbeatInterval,
	)

	var leaderCount int
	var currentLeader *Node

	for _, n := range nodes {
		snapshot := n.RaftSnapshot()

		if snapshot.Role == raft.RoleLeader {
			leaderCount++
			currentLeader = n
		}
	}

	if leaderCount != 1 {
		t.Fatalf(
			"expected exactly one stable leader, got %d",
			leaderCount,
		)
	}

	if currentLeader.ID != leader.ID {
		t.Fatalf(
			"leader changed from %s to %s",
			leader.ID,
			currentLeader.ID,
		)
	}

	current := currentLeader.RaftSnapshot()

	if current.CurrentTerm != first.CurrentTerm {
		t.Fatalf(
			"term changed from %d to %d while cluster was healthy",
			first.CurrentTerm,
			current.CurrentTerm,
		)
	}
}

func TestLeaderReplicatesAndCommitsEntry(
	t *testing.T,
) {
	nodes := newThreeNodeTestCluster(t)

	startTestCluster(t, nodes)

	leader, snapshot := waitForLeader(
		t,
		nodes,
		5*time.Second,
	)

	entry, err :=
		leader.raftState.AppendLeaderCommand(
			command.Command{
				Type:  command.TypeSet,
				Key:   "name",
				Value: "alice",
			},
		)

	if err != nil {
		t.Fatal(err)
	}

	leader.sendHeartbeatRound(
		snapshot.CurrentTerm,
	)

	deadline :=
		time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		allReplicated := true

		for _, n := range nodes {
			entries :=
				n.raftState.LogEntries()

			if len(entries) < 1 {
				allReplicated = false
				break
			}
		}

		if allReplicated {
			break
		}

		time.Sleep(
			25 * time.Millisecond,
		)
	}

	for _, n := range nodes {
		entries :=
			n.raftState.LogEntries()

		if len(entries) != 1 {
			t.Fatalf(
				"node %s expected 1 log entry, got %d",
				n.ID,
				len(entries),
			)
		}

		if entries[0].Command.Key != "name" ||
			entries[0].Command.Value != "alice" {
			t.Fatalf(
				"node %s has unexpected command",
				n.ID,
			)
		}
	}

	got := leader.RaftSnapshot()

	if got.CommitIndex < entry.Index {
		t.Fatalf(
			"expected entry %d committed, commitIndex=%d",
			entry.Index,
			got.CommitIndex,
		)
	}
}
