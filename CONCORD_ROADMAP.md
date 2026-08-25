# Concord — Accelerated First-Principles Build Roadmap

> **Deadline:** September 2, 2026
>
> **Current status:** Steps **1–32 complete**. Continue from **Step 33**.
>
> **Purpose:** This file is the source of truth for continuing Concord with ChatGPT or another AI. It replaces the original roadmap numbering, which no longer matched the implementation order.

---

## 0. Current snapshot

- Project: **Concord**, a distributed replicated key-value store written in Go.
- Current completion estimate: **~20% by implementation difficulty**, not by raw step count.
- Current step: **33**.
- Current working milestone: a concurrent TCP node with an in-memory KV store and networked `SET`, `GET`, and `DELETE`.
- `SET` and `DELETE` wait for an explicit `OK` acknowledgement.
- `GET` receives `VALUE` or `NOT_FOUND`.
- Request/response operations use network deadlines.
- Multiple peers can be configured by stable node ID and targeted with `--target`.
- Three nodes can be run locally and communicate.
- Current major correctness debt: `Store` uses an unprotected Go map while connection handlers run concurrently.
- Hard v1 target: **working three-node Raft replication + crash recovery + failover demo by September 2, 2026**.

### Repository shape as of the Step 32 checkpoint

```text
Concord/
├── cmd/
│   └── concord/
│       └── main.go
├── internal/
│   ├── config/
│   ├── node/
│   ├── peer/
│   ├── protocol/
│   └── store/
├── go.mod
└── CONCORD_ROADMAP.md
```

Expected new packages later include `internal/wal` and Raft-related code. Create them only when the corresponding roadmap step is reached.

---

## 1. Rules for any AI helping with Concord

Concord is primarily a **learning project**. Do not turn this roadmap into a code-generation sprint that hides the concepts.

### Teaching rules

- Assume beginner-level knowledge unless a completed step proves the concept has already been learned.
- Explain **why** a concept exists before implementing it.
- Prefer intuition and observable experiments before formalism.
- Work **one numbered step at a time** unless the user explicitly asks to batch steps.
- Tell the user exactly which file is changing.
- Keep each code change small enough to understand.
- Explain non-obvious Go, networking, OS, storage, and distributed-systems behavior.
- After meaningful changes, run or ask the user to run tests/experiments.
- When something fails: read the error, form a hypothesis, test it, then make the smallest fix.
- Prefer the Go standard library while the project is teaching fundamentals.
- Never hide Raft behind a consensus library.
- Do not silently skip correctness issues just to hit the deadline.
- Reuse concepts already learned; do not re-explain familiar syntax at the same depth every time.

### Deadline rule

The September 2 deadline changes **prioritization**, not correctness. The core must be finished before optional polish.

Priority order:

1. Concurrency correctness.
2. Persistence / WAL / recovery.
3. Raft leader election.
4. Raft log replication and quorum commit.
5. Crash/failure behavior and integration tests.
6. Demo/observability.
7. Documentation/polish.

If behind schedule, cut or postpone in this order:

1. Snapshots/log compaction.
2. Dynamic cluster membership.
3. Fancy dashboard styling.
4. Public cloud deployment if it is unreliable or time-consuming.
5. Extra metrics/benchmarks.

Do **not** cut Raft correctness, durable state, race testing, leader failover, or the three-node demo.

---

## 2. What Concord v1 must do by September 2

```text
Client
  |
  v
+----------------+
| Raft Leader    |
| Concord node 1 |
+-------+--------+
        | AppendEntries
   +----+----+
   |         |
   v         v
+------+   +------+
|node 2|   |node 3|
|foll. |   |foll. |
+------+   +------+
```

Required v1 behavior:

- Run a three-node cluster.
- Elect exactly one leader in normal operation.
- Accept `SET`, `GET`, and `DELETE` through the leader.
- Replicate mutating commands through a Raft log.
- Acknowledge a write only after quorum commit.
- Apply committed commands to the in-memory state machine.
- Persist enough state/log data to recover safely after restart.
- Elect a new leader when the current leader fails.
- Catch up a restarted follower.
- Pass unit/integration tests and the Go race detector.
- Provide a recruiter-friendly demo and strong README/architecture explanation.

Not required for September 2 v1: snapshots, log compaction, dynamic membership, TLS, production-grade security, or production-scale performance.

---

## 3. Known cleanup/debt at Step 32

- `Store.data` is a plain `map[string]string`; concurrent handlers can race.
- Config currently contains duplicated `--ping` + `--message` validation.
- `PingTimeout` is now being reused as a generic request timeout and should be renamed/generalized.
- `Ping`, `Get`, and `SendAndWaitForOK` contain intentional request/response duplication that can be cleaned up carefully.
- Current storage is memory-only.
- Current nodes communicate but do **not** replicate KV state.
- There is no leader/follower role yet.
- There is no graceful shutdown or structured observability yet.

---

## 4. Daily deadline plan

| Day | Date | Steps | Milestone | Cumulative target |
|---|---|---:|---|---:|
| Day 1 | Aug 22 | 1–23 | Go/TCP foundations, structured messages, peers, 3-node communication, command/server modes | ~10% |
| Day 2 | Aug 25 | 24–32 | Tests, in-memory KV store, SET/GET/DELETE over TCP, request/response, write ACKs | ~20% |
| Day 3 | Aug 26 | 33–50 | Concurrency safety + WAL + restart recovery | ~32% |
| Day 4 | Aug 27 | 51–70 | Raft state, terms, voting, first leader election | ~43% |
| Day 5 | Aug 28 | 71–95 | Stable elections + heartbeats + leader failover | ~57% |
| Day 6 | Aug 29 | 96–125 | Raft log replication + quorum commit + state-machine apply | ~72% |
| Day 7 | Aug 30 | 126–150 | Partition safety, leadership changes, concurrent replicated writes | ~82% |
| Day 8 | Aug 31 | 151–170 | Crash recovery, checksummed WAL, follower rejoin, integration tests | ~90% |
| Day 9 | Sep 1 | 171–188 | HTTP status/API, dashboard, Docker Compose, demo rehearsal | ~96% |
| Day 10 | Sep 2 | 189–200 | README, architecture docs, final testing, demo/release | 100% |

Percentages are approximate by **difficulty and project value**, not by number of checkboxes.

---

## 5. Reconciled completed steps — 1 through 32

> Steps 1–32 below describe the implementation milestones actually reached. The earliest micro-order was reconstructed from the current repository and project history; the resulting code state is what matters for continuation.

### Step 1 — COMPLETE
- [x] Initialize the Go module and Concord CLI skeleton.

### Step 2 — COMPLETE
- [x] Add basic node configuration: node ID and listening port.

### Step 3 — COMPLETE
- [x] Create the Node type and constructor.

### Step 4 — COMPLETE
- [x] Start a TCP listener for a Concord node.

### Step 5 — COMPLETE
- [x] Accept incoming TCP connections.

### Step 6 — COMPLETE
- [x] Read data from an incoming TCP connection.

### Step 7 — COMPLETE
- [x] Handle connections concurrently with goroutines.

### Step 8 — COMPLETE
- [x] Introduce a Peer type for other Concord nodes.

### Step 9 — COMPLETE
- [x] Parse repeatable peer configuration from the CLI.

### Step 10 — COMPLETE
- [x] Open outgoing TCP connections to peers.

### Step 11 — COMPLETE
- [x] Introduce protocol.Message for structured messages.

### Step 12 — COMPLETE
- [x] Encode/decode messages as JSON.

### Step 13 — COMPLETE
- [x] Use newline-delimited framing for TCP messages.

### Step 14 — COMPLETE
- [x] Send structured messages between nodes.

### Step 15 — COMPLETE
- [x] Add the PING request.

### Step 16 — COMPLETE
- [x] Add the PONG response on the same TCP connection.

### Step 17 — COMPLETE
- [x] Add network deadlines/timeouts for request-response operations.

### Step 18 — COMPLETE
- [x] Give peers stable IDs instead of addressing them only by host:port.

### Step 19 — COMPLETE
- [x] Run and communicate across a local three-node cluster.

### Step 20 — COMPLETE
- [x] Validate peer configuration: duplicate IDs/addresses and self-peers.

### Step 21 — COMPLETE
- [x] Target a peer by ID with --target.

### Step 22 — COMPLETE
- [x] Support one-shot client commands such as ping/message.

### Step 23 — COMPLETE
- [x] Separate one-shot command mode from long-running server mode.

### Step 24 — COMPLETE
- [x] Add protocol round-trip tests.

### Step 25 — COMPLETE
- [x] Add configuration/peer validation tests.

### Step 26 — COMPLETE
- [x] Add node constructor/peer lookup tests and keep go test ./... green.

### Step 27 — COMPLETE
- [x] Create the in-memory Store with Set/Get/Delete and unit tests.

### Step 28 — COMPLETE
- [x] Make every Node own an independent Store and test that isolation.

### Step 29 — COMPLETE
- [x] Implement SET over TCP and store the value on the target node.

### Step 30 — COMPLETE
- [x] Implement GET over TCP with VALUE and NOT_FOUND responses.

### Step 31 — COMPLETE
- [x] Implement DELETE over TCP and verify the full SET/GET/DELETE lifecycle.

### Step 32 — COMPLETE
- [x] Add OK acknowledgements for SET/DELETE and test SendAndWaitForOK.

---

# FUTURE PLAN — Continue from Step 33

## Day 3 — Wednesday, August 26 — Concurrency + persistence foundations

### Step 33
- [ ] Run the race detector and understand why Store is unsafe under concurrent handlers.

### Step 34
- [ ] Learn mutex vs RWMutex and decide the Store locking policy.

### Step 35
- [ ] Add an RWMutex to Store.

### Step 36
- [ ] Protect Store.Set with a write lock.

### Step 37
- [ ] Protect Store.Get with a read lock.

### Step 38
- [ ] Protect Store.Delete with a write lock.

### Step 39
- [ ] Add concurrent Store tests.

### Step 40
- [ ] Run go test -race ./... and make the current code race-clean.

### Step 41
- [ ] Clean the duplicate --ping/--message validation in config.

### Step 42
- [ ] Rename/generalize PingTimeout into a request timeout without changing behavior.

### Step 43
- [ ] Identify duplication in Ping/Get/SendAndWaitForOK; refactor only the safe common pieces.

### Step 44
- [ ] Define the first append-only WAL record for a key-value command.

### Step 45
- [ ] Create internal/wal and open/create a log file.

### Step 46
- [ ] Append newline-delimited WAL records.

### Step 47
- [ ] Flush/sync WAL writes and understand durability vs performance.

### Step 48
- [ ] Write SET/DELETE to WAL before mutating the Store.

### Step 49
- [ ] Replay WAL records during startup to rebuild in-memory state.

### Step 50
- [ ] Add restart-recovery tests and manually prove data survives a node restart.

**End-of-day acceptance criterion:** Store is race-safe, WAL-backed, and data survives a restart.

---

## Day 4 — Thursday, August 27 — Raft state + first leader election

### Step 51
- [ ] Explain the exact problem Raft solves: agreement despite failures.

### Step 52
- [ ] Define Follower, Candidate, and Leader roles.

### Step 53
- [ ] Create the initial Raft state: role, currentTerm, votedFor.

### Step 54
- [ ] Attach Raft state to a Concord node without changing KV behavior yet.

### Step 55
- [ ] Protect Raft state against concurrent access.

### Step 56
- [ ] Define RequestVote request/response protocol data.

### Step 57
- [ ] Handle an incoming RequestVote RPC.

### Step 58
- [ ] Reject vote requests from stale terms.

### Step 59
- [ ] Update local term and step down when a higher term is observed.

### Step 60
- [ ] Implement one-vote-per-term logic.

### Step 61
- [ ] Persist currentTerm and votedFor so a restart cannot violate voting safety.

### Step 62
- [ ] Load persistent Raft metadata during startup.

### Step 63
- [ ] Learn election timeouts and why they are randomized.

### Step 64
- [ ] Create a randomized election timeout.

### Step 65
- [ ] Start/reset the follower election timer.

### Step 66
- [ ] Transition Follower -> Candidate when the timeout fires.

### Step 67
- [ ] Increment term and vote for self when becoming Candidate.

### Step 68
- [ ] Send RequestVote RPCs to peers concurrently.

### Step 69
- [ ] Count votes using majority/quorum logic.

### Step 70
- [ ] Transition Candidate -> Leader after receiving a majority.

**End-of-day acceptance criterion:** A three-node cluster can perform its first real Raft leader election.

---

## Day 5 — Friday, August 28 — Heartbeats + stable elections

### Step 71
- [ ] Unit-test majority calculation.

### Step 72
- [ ] Test stale-term vote rejection.

### Step 73
- [ ] Test that a node cannot vote twice in one term.

### Step 74
- [ ] Run the first three-node leader-election experiment.

### Step 75
- [ ] Define an empty AppendEntries RPC to act as a heartbeat.

### Step 76
- [ ] Start a periodic heartbeat loop when a node becomes Leader.

### Step 77
- [ ] Reset follower election timeout after a valid heartbeat.

### Step 78
- [ ] Step down when a heartbeat carries a higher term.

### Step 79
- [ ] Make a Candidate step down after valid AppendEntries from the current/higher term.

### Step 80
- [ ] Make a Leader step down if it discovers a higher term.

### Step 81
- [ ] Tune and document heartbeat/election timeout relationships.

### Step 82
- [ ] Prevent overlapping election rounds from corrupting state.

### Step 83
- [ ] Add cancellation/lifecycle control for election and heartbeat goroutines.

### Step 84
- [ ] Make Raft Start/Stop behavior safe and understandable.

### Step 85
- [ ] Track the currently known leader ID.

### Step 86
- [ ] Expose a small local status view: role, term, leader.

### Step 87
- [ ] Improve logs around role/term transitions.

### Step 88
- [ ] Kill the current leader manually and observe a new election.

### Step 89
- [ ] Verify the remaining two nodes can elect a leader.

### Step 90
- [ ] Restart the old leader and verify it rejoins as a follower.

### Step 91
- [ ] Add an automated three-node election integration test.

### Step 92
- [ ] Handle vote RPC timeouts without blocking an election forever.

### Step 93
- [ ] Test election behavior when one peer is unavailable.

### Step 94
- [ ] Run the race detector against election tests and fix discovered races.

### Step 95
- [ ] Checkpoint: repeatedly elect exactly one stable leader in a healthy three-node cluster.

**End-of-day acceptance criterion:** Leader elections are stable; leader failure causes a new election; old leader rejoins safely.

---

## Day 6 — Saturday, August 29 — Raft log replication

### Step 96
- [ ] Define LogEntry with index, term, and command.

### Step 97
- [ ] Add a Raft log to each node.

### Step 98
- [ ] Only allow the Leader to propose mutating client commands.

### Step 99
- [ ] Define a NOT_LEADER response for writes sent to followers.

### Step 100
- [ ] Include known leader information in NOT_LEADER when available.

### Step 101
- [ ] Extend AppendEntries with prevLogIndex, prevLogTerm, entries, and leaderCommit.

### Step 102
- [ ] Define AppendEntries success/failure response data.

### Step 103
- [ ] Implement the follower prev-log consistency check.

### Step 104
- [ ] Reject AppendEntries when prevLogIndex/prevLogTerm do not match.

### Step 105
- [ ] Detect conflicting follower entries.

### Step 106
- [ ] Delete a conflicting suffix before appending leader entries.

### Step 107
- [ ] Append new entries on the follower.

### Step 108
- [ ] Track nextIndex for each follower on the Leader.

### Step 109
- [ ] Track matchIndex for each follower on the Leader.

### Step 110
- [ ] Initialize replication indices when leadership begins.

### Step 111
- [ ] Send a newly proposed log entry to one follower.

### Step 112
- [ ] Retry replication after a log mismatch.

### Step 113
- [ ] Back up nextIndex until the logs line up.

### Step 114
- [ ] Replicate an entry to both followers concurrently.

### Step 115
- [ ] Compute whether an entry exists on a majority.

### Step 116
- [ ] Advance commitIndex only using Raft's current-term commit rule.

### Step 117
- [ ] Apply a committed SET entry to Store.

### Step 118
- [ ] Apply a committed DELETE entry to Store.

### Step 119
- [ ] Serve GET from the Leader's applied state; reject/redirect follower client reads for v1.

### Step 120
- [ ] Make client SET wait for the log entry to commit before replying OK.

### Step 121
- [ ] Make client DELETE wait for commit before replying OK.

### Step 122
- [ ] Return a timeout/error when a write cannot reach quorum.

### Step 123
- [ ] Catch up a follower that missed several entries.

### Step 124
- [ ] Add log consistency and replication unit tests.

### Step 125
- [ ] Checkpoint: a client write is replicated and committed across a healthy three-node cluster.

**End-of-day acceptance criterion:** A leader replicates and commits SET/DELETE through a majority before acknowledging success.

---

## Day 7 — Sunday, August 30 — Commit safety + partitions + concurrency

### Step 126
- [ ] Persist Raft log entries using the WAL foundation.

### Step 127
- [ ] Include index and term in durable log records.

### Step 128
- [ ] Reconstruct the Raft log after restart.

### Step 129
- [ ] Test a follower restart followed by catch-up.

### Step 130
- [ ] Test a leader restart and subsequent re-election.

### Step 131
- [ ] Construct a divergent-log scenario intentionally.

### Step 132
- [ ] Verify conflicting uncommitted entries can be overwritten.

### Step 133
- [ ] Verify committed entries are never overwritten.

### Step 134
- [ ] Partition one follower from the other two nodes.

### Step 135
- [ ] Verify the majority side can continue committing writes.

### Step 136
- [ ] Partition the current Leader into the minority side.

### Step 137
- [ ] Verify the majority side elects a new Leader.

### Step 138
- [ ] Verify the isolated old Leader cannot successfully commit new writes.

### Step 139
- [ ] Heal the partition.

### Step 140
- [ ] Verify logs converge after healing.

### Step 141
- [ ] Verify stale leaders/candidates step down when they observe higher terms.

### Step 142
- [ ] Add an integration test for commit safety across a leadership change.

### Step 143
- [ ] Send multiple client writes concurrently.

### Step 144
- [ ] Serialize/protect proposal and log state correctly.

### Step 145
- [ ] Track pending client proposals waiting for commit.

### Step 146
- [ ] Wake the correct client only when its entry commits.

### Step 147
- [ ] Add command/request IDs if needed to correlate pending operations safely.

### Step 148
- [ ] Document v1 duplicate-request/idempotency limitations instead of hiding them.

### Step 149
- [ ] Add a concurrent-write three-node integration test.

### Step 150
- [ ] Checkpoint: go test -race ./... passes with replicated concurrent writes.

**End-of-day acceptance criterion:** Partitions and leadership changes preserve committed data; concurrent writes are race-clean.

---

## Day 8 — Monday, August 31 — Crash recovery + failure hardening

### Step 151
- [ ] Handle SIGINT/SIGTERM for graceful shutdown.

### Step 152
- [ ] Close the TCP listener cleanly.

### Step 153
- [ ] Stop election/heartbeat/replication goroutines cleanly.

### Step 154
- [ ] Close/sync WAL files during shutdown.

### Step 155
- [ ] Replay durable Raft state at startup.

### Step 156
- [ ] Understand what a partially written WAL tail looks like after a crash.

### Step 157
- [ ] Add a checksum to WAL records.

### Step 158
- [ ] Verify checksums while reading WAL.

### Step 159
- [ ] Safely ignore/recover from an incomplete final WAL record.

### Step 160
- [ ] Fail loudly on corruption in a non-tail WAL record.

### Step 161
- [ ] Document the exact fsync/durability guarantee Concord v1 provides.

### Step 162
- [ ] Add crash/restart tests for persisted terms, votes, and log entries.

### Step 163
- [ ] Crash a node after writes and prove recovery manually.

### Step 164
- [ ] Keep one follower offline while writes occur, then restart it.

### Step 165
- [ ] Verify the restarted follower catches up.

### Step 166
- [ ] Create a repeatable local partition/failure helper script or test harness.

### Step 167
- [ ] Create reusable integration-test helpers for starting three nodes.

### Step 168
- [ ] Make integration timeouts deterministic enough for reliable tests.

### Step 169
- [ ] Add a full leader-failure -> re-election -> continued-writes integration test.

### Step 170
- [ ] Checkpoint: demonstrate crash recovery and follower rejoin without losing committed data.

**End-of-day acceptance criterion:** Committed data survives crashes/restarts and lagging followers catch up.

---

## Day 9 — Tuesday, September 1 — Recruiter-visible demo + observability

### Step 171
- [ ] Add a small HTTP status server without replacing the TCP/Raft transport.

### Step 172
- [ ] Expose node ID, role, term, leader, commit index, and last log index as JSON.

### Step 173
- [ ] Add a health endpoint.

### Step 174
- [ ] Add simple counters for requests, elections, commits, and replication failures.

### Step 175
- [ ] Expose useful metrics/status without introducing a heavy framework.

### Step 176
- [ ] Make operational logs consistent enough to follow elections and commits.

### Step 177
- [ ] Create a minimal web dashboard shell.

### Step 178
- [ ] Display the three nodes and their current roles.

### Step 179
- [ ] Visually identify the current Leader and current term.

### Step 180
- [ ] Add HTTP endpoints/forms for demo SET, GET, and DELETE operations.

### Step 181
- [ ] Wire dashboard controls to the HTTP API.

### Step 182
- [ ] Poll/refresh cluster status so leadership changes become visible.

### Step 183
- [ ] Create a Dockerfile for Concord.

### Step 184
- [ ] Create a Docker Compose three-node demo cluster.

### Step 185
- [ ] Run the full cluster through Docker Compose.

### Step 186
- [ ] Create a demo action/script to kill the current Leader.

### Step 187
- [ ] Restart a killed node and show it rejoining/catching up.

### Step 188
- [ ] Checkpoint: rehearse the complete recruiter demo from a clean checkout.

**End-of-day acceptance criterion:** A clean-checkout recruiter demo can show cluster state, KV operations, leader death, failover, and recovery.

---

## Day 10 — Wednesday, September 2 — Documentation + final v1 release

### Step 189
- [ ] Write README quickstart instructions from a clean machine/check-out perspective.

### Step 190
- [ ] Add a clear architecture diagram.

### Step 191
- [ ] Document the TCP/application protocol and important message types.

### Step 192
- [ ] Explain Concord's Raft implementation in plain English.

### Step 193
- [ ] Document guarantees, assumptions, and known limitations honestly.

### Step 194
- [ ] Document unit, integration, race-detector, and failure-test commands.

### Step 195
- [ ] Write the exact local three-node demo procedure.

### Step 196
- [ ] Choose the final demo delivery: public deployment if reliable, otherwise Docker/local cluster plus recorded failover demo.

### Step 197
- [ ] Record/rehearse the leader-failure and recovery demo.

### Step 198
- [ ] Run gofmt, go vet, go test ./..., and go test -race ./...; fix release-blocking failures.

### Step 199
- [ ] Final code/config cleanup, meaningful final commits, and optionally tag v1.0.0.

### Step 200
- [ ] Final audit: clean checkout works, three-node cluster elects, writes replicate/commit, leader fails over, recovery works, docs/demo are complete.

**End-of-day acceptance criterion:** Concord v1 passes the final audit and is ready to show.

---

## 6. Testing discipline

Run frequently:

```bash
go test ./...
```

Once concurrency begins, also run:

```bash
go test -race ./...
```

Before the final release:

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
```

Integration tests should prove behavior, not merely code coverage. Important scenarios include:

- healthy three-node election
- leader failure and re-election
- one follower unavailable
- old leader rejoining
- lagging follower catch-up
- majority/minority partition behavior
- concurrent client writes
- crash/restart with durable state
- divergent uncommitted logs converging after healing

---

## 7. Git checkpoint guidance

Commit after coherent learning/behavior milestones, not after every tiny edit.

Good examples:

```text
make store concurrency safe
add WAL persistence and recovery
add raft voting state
implement leader election
add raft heartbeats
replicate raft log entries
commit commands through quorum
add crash recovery tests
add cluster dashboard
document concord architecture
```

Before committing:

```bash
git diff
go test ./...
```

And after Step 33 onward, use the race detector at relevant concurrency checkpoints.

---

## 8. What another AI should do when handed this file

1. Read this file before proposing implementation work.
2. Read the current repository state; do not assume the file is newer than the code.
3. Find the first unchecked numbered step.
4. Continue from that step only.
5. Preserve the teaching rules.
6. Do not dump an entire Raft implementation at once.
7. Explain important invariants before coding them.
8. Keep the September 2 deadline visible when prioritizing work.
9. At the end of each session, update the Progress Tracker below.

---

## 9. Progress Tracker

```text
Deadline: September 2, 2026
Current date at roadmap rewrite: August 25, 2026
Current step: 33
Completed through: Step 32
Approximate completion: 20%

Current milestone:
Networked in-memory KV database with SET/GET/DELETE, TCP request-response, deadlines, and write acknowledgements.

Next milestone:
Concurrency-safe Store, race-detector clean build, WAL persistence, and restart recovery.

Next command/experiment:
go test -race ./...
```

### Completed major capabilities

- [x] Go module and CLI config
- [x] TCP server/client communication
- [x] concurrent connection handlers
- [x] structured JSON + newline framing
- [x] named peer configuration and targeting
- [x] PING/PONG and deadlines
- [x] local three-node communication
- [x] in-memory Store
- [x] SET over TCP
- [x] GET with VALUE / NOT_FOUND
- [x] DELETE over TCP
- [x] OK acknowledgements for writes
- [x] unit tests for current core pieces
- [ ] race-safe Store
- [ ] WAL persistence / crash recovery
- [ ] Raft roles/terms/voting
- [ ] leader election
- [ ] heartbeats
- [ ] log replication
- [ ] quorum commit
- [ ] follower catch-up
- [ ] partition/failure integration tests
- [ ] recruiter dashboard/demo
- [ ] final documentation/release

---

## 10. Post-September-2 stretch roadmap

Only start these after v1 is complete:

- snapshots and log compaction
- dynamic cluster membership
- read-index / stronger linearizable-read mechanisms
- client retry/idempotency improvements
- richer metrics and tracing
- TLS/authentication
- benchmarks and performance tuning
- more sophisticated chaos testing
- public multi-node deployment if it is worth maintaining

---

## Final definition of done for Concord v1

Concord v1 is done when a fresh checkout can be used to demonstrate:

```text
1. Start three nodes
2. Observe one leader
3. SET a key through the leader
4. GET the committed value
5. Prove the write exists on a quorum / replicated log
6. Kill the leader
7. Observe a new leader election
8. Continue writing successfully
9. Restart the failed node
10. Observe it catch up
11. Run tests + race detector successfully
12. Explain why the system remained safe
```

If those twelve things work and the README explains the architecture and limitations clearly, the September 2 goal has been met.