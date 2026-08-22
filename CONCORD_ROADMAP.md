# Concord — First-Principles Build Roadmap

> **Purpose of this file**
>
> This document is the source of truth for building **Concord**.
>
> If I lose access to the original ChatGPT conversation, I should be able to give this file to another AI and say:
>
> **"Continue Concord from Step X. Follow the teaching rules in this document."**
>
> The AI should then be able to continue without assuming I already know the underlying concepts.

---

# 0. Who is building this?

I am building Concord primarily to **learn**.

My main learning goals are:

1. Learn **Go** deeply.
2. Learn **distributed systems** from first principles.
3. Learn **network programming**.
4. Learn **concurrency** and synchronization.
5. Learn **storage systems**.
6. Learn **consensus algorithms**, especially Raft.
7. Learn how real backend systems are structured.
8. Learn testing, debugging, observability, deployment, and failure analysis.
9. End with a project strong enough to discuss seriously in interviews.
10. Be able to explain every important line of code I write.

I do **not** want the project generated for me all at once.

I will manually create the folders and files.

I will manually type or copy small pieces of code only after understanding them.

---

# 1. Rules for any AI helping with Concord

These rules are extremely important.

## Teaching rules

The assistant must:

- Assume I am a beginner unless this file explicitly says I have already learned something.
- Explain unfamiliar terminology before using it heavily.
- Build from first principles.
- Prefer intuition before formal definitions.
- Explain **why** something exists before showing how to implement it.
- Explain what problem each new abstraction solves.
- Keep implementation steps small.
- Avoid dumping entire subsystems at once.
- Ask me to reason about code sometimes instead of immediately giving the answer.
- Point out important Go language concepts as they appear.
- Point out important operating-system and networking concepts as they appear.
- Point out distributed-systems failure modes when relevant.
- Explain important trade-offs rather than pretending there is one perfect design.
- Distinguish educational shortcuts from production-grade techniques.
- Revisit earlier design decisions when the project becomes sophisticated enough to improve them.

## Coding rules

Unless I explicitly request otherwise:

- Work on **one roadmap step at a time**.
- Do not skip steps silently.
- Do not create 10 files at once.
- Tell me exactly which file we are changing.
- Explain the purpose of the file before writing code.
- Keep code changes small enough that I can understand them.
- Explain every non-obvious line.
- After meaningful code changes, tell me how to run or test them.
- Make me observe the behavior rather than merely saying "it works."
- Prefer the Go standard library when learning value is higher.
- Introduce third-party libraries only when they solve a problem we already understand.
- Avoid magical frameworks.
- Do not hide important distributed-systems behavior behind a library.

## Debugging rules

When something fails:

1. Do not immediately replace the code.
2. Read the error.
3. Explain what the error means.
4. Form a hypothesis.
5. Test the hypothesis.
6. Fix the smallest underlying issue.
7. Explain why the fix works.

## Project-management rule

At the end of a session, update the **Progress Tracker** section in this file if possible.

---

# 2. What is Concord?

**Concord is a distributed, replicated key-value store written in Go.**

At the simplest level, a client should eventually be able to do things like:

```text
SET name alice
GET name
DELETE name
```

But Concord will not remain a single-process toy database.

We will gradually turn it into a real distributed system.

A Concord cluster will eventually contain multiple nodes:

```text
          Client
             |
             v
      +-------------+
      |   Leader    |
      | Concord #1  |
      +-------------+
        /         \
       /           \
      v             v
+-------------+ +-------------+
|  Follower   | |  Follower   |
| Concord #2  | | Concord #3  |
+-------------+ +-------------+
```

The nodes will replicate commands using the **Raft consensus algorithm**.

If the current leader dies, the remaining nodes should elect a new leader.

The system will eventually include:

- TCP networking
- an application protocol
- client/server architecture
- concurrent connections
- an in-memory key-value store
- persistent storage
- a write-ahead log
- checksums
- crash recovery
- replicated logs
- leader election
- heartbeats
- Raft terms
- Raft voting
- quorum / majority logic
- log replication
- commit indexes
- state-machine application
- snapshots
- log compaction
- cluster membership
- graceful shutdown
- metrics
- structured logging
- health endpoints
- debugging tooling
- integration tests
- chaos/failure tests
- Docker
- deployment/demo modes
- a small frontend/dashboard
- documentation and architecture diagrams

---

# 3. What Concord is NOT

At least initially, Concord is **not**:

- PostgreSQL
- Redis
- etcd
- Consul
- ZooKeeper
- a SQL database
- a production-ready database
- a Kubernetes replacement
- a blockchain
- a peer-to-peer system

It is an educational distributed system inspired by ideas used in systems such as etcd and Consul.

---

# 4. Final learning architecture

The target architecture is approximately:

```text
                           +----------------------+
                           |   Web Dashboard      |
                           | cluster visualization|
                           +----------+-----------+
                                      |
                                      v
                           +----------------------+
                           |   Concord HTTP API   |
                           +----------+-----------+
                                      |
                                      v
                +---------------------------------------------+
                |                Concord Node                  |
                |                                             |
 Client ------> |  Client Server / Command Handler            |
                |                 |                           |
                |                 v                           |
                |          +-------------+                    |
                |          | Raft Layer  |                    |
                |          +------+------+                    |
                |                 |                           |
                |         replicated log                      |
                |                 |                           |
                |                 v                           |
                |          +-------------+                    |
                |          | State       |                    |
                |          | Machine     |                    |
                |          +------+------+                    |
                |                 |                           |
                |                 v                           |
                |          Key/Value Store                    |
                |                                             |
                | Persistence: WAL + snapshots                |
                +---------------------------------------------+
                    ^                ^                ^
                    |                |                |
                    +-------- node-to-node RPC -------+
```

Do not build this all at once.

We will earn each layer gradually.

---

# 5. Technologies

Primary language:

- **Go**

Likely supporting technologies later:

- Go standard library
- TCP
- HTTP
- JSON initially where useful
- custom message formats where educational
- Docker
- Docker Compose
- Git
- GitHub
- Prometheus-style metrics
- optional lightweight HTML/CSS/JavaScript frontend

We should avoid unnecessary dependencies early.

---

# 6. Repository target structure

This is a **future target**, not something to create immediately.

```text
concord/
├── cmd/
│   ├── concord/
│   └── concordctl/
├── internal/
│   ├── server/
│   ├── protocol/
│   ├── store/
│   ├── wal/
│   ├── raft/
│   ├── transport/
│   ├── snapshot/
│   └── observability/
├── web/
├── tests/
├── scripts/
├── docs/
├── deployments/
├── go.mod
├── go.sum
├── README.md
└── ROADMAP.md
```

We will create directories only when they become necessary.

---

# 7. Mental model of the build

We will evolve Concord through these major versions:

```text
V0  -> hello-world Go program
V1  -> local in-memory key-value store
V2  -> single-node TCP server
V3  -> concurrent server
V4  -> persistent single-node database
V5  -> manually replicated multi-node system
V6  -> Raft leader election
V7  -> Raft log replication
V8  -> crash-tolerant replicated key-value store
V9  -> snapshots and membership
V10 -> observability, dashboard, deployment, chaos demo
```

The point is not reaching V10 quickly.

The point is understanding every transition.

---

# 8. Roadmap legend

Each roadmap item has a checkbox.

```text
[ ] not started
[x] completed
[~] currently working on it
```

When resuming with another AI, say:

> Read `CONCORD_ROADMAP.md`. We have completed through Step X. Continue with Step X+1 and obey the teaching rules.

---

# PHASE 0 — Environment, Go basics, and project foundations

## Goal

Understand what we are running before building anything distributed.

### Step 1 — Define the project in our own words

- [ ] Explain what a key-value store is.
- [ ] Explain what "distributed" means.
- [ ] Explain what replication means.
- [ ] Explain at a high level why consensus is needed.
- [ ] Write a two-sentence description of Concord ourselves.

**Learning checkpoint:** I should be able to explain Concord without using buzzwords.

---

### Step 2 — Verify the Go installation

- [ ] Run `go version`.
- [ ] Understand what the Go toolchain is.
- [ ] Understand compiler vs runtime at a beginner level.
- [ ] Understand what `go run` does.
- [ ] Understand what `go build` does.

---

### Step 3 — Create the repository manually

- [ ] Create a directory named `concord`.
- [ ] Enter it.
- [ ] Run `git init`.
- [ ] Understand what a Git repository actually stores.
- [ ] Check `git status`.

Do not add application code yet.

---

### Step 4 — Create the Go module

- [ ] Run `go mod init ...`.
- [ ] Understand what a Go module is.
- [ ] Open `go.mod`.
- [ ] Explain every line currently inside `go.mod`.

---

### Step 5 — Write the smallest possible Go program

Create:

```text
main.go
```

- [ ] Write `package main`.
- [ ] Import `fmt`.
- [ ] Create `func main()`.
- [ ] Print `"Concord starting..."`.
- [ ] Run it.
- [ ] Explain packages, imports, functions, and strings.

---

### Step 6 — Build a binary

- [ ] Run `go build`.
- [ ] Locate the produced executable.
- [ ] Run the executable directly.
- [ ] Understand source code vs compiled binary.
- [ ] Understand why Go binaries are convenient for servers.

---

### Step 7 — Learn variables and basic types through Concord config

Create a tiny configuration inside `main.go`:

- [ ] string
- [ ] integer
- [ ] boolean
- [ ] short variable declaration `:=`
- [ ] explicit `var`
- [ ] `const`

Use realistic examples such as node ID and port.

---

### Step 8 — Learn functions

- [ ] Move a small operation into a function.
- [ ] Learn parameters.
- [ ] Learn return values.
- [ ] Learn multiple return values.
- [ ] Learn why Go commonly returns `(value, error)`.

---

### Step 9 — Learn structs

Create a tiny `Node` struct.

Potential fields:

```text
ID
Address
```

- [ ] Understand what a struct is.
- [ ] Instantiate one.
- [ ] Read its fields.
- [ ] Modify a field.

---

### Step 10 — Learn methods and pointers

- [ ] Add a method to `Node`.
- [ ] Understand a value receiver.
- [ ] Understand a pointer receiver.
- [ ] Learn what `&` and `*` mean.
- [ ] Explain why server objects often use pointers.

---

### Step 11 — Learn slices and maps

- [ ] Create a slice of peer addresses.
- [ ] Create a map representing key/value data.
- [ ] Add entries.
- [ ] Read entries.
- [ ] Delete entries.
- [ ] Understand zero values and the `value, ok` map pattern.

---

### Step 12 — Learn errors

- [ ] Create an error with `errors.New`.
- [ ] Return an error from a function.
- [ ] Check `if err != nil`.
- [ ] Understand why errors are values in Go.
- [ ] Avoid `panic` for ordinary failures.

---

### Step 13 — Learn packages by refactoring

Create only when needed:

```text
internal/store/
```

- [ ] Move key-value logic into a package.
- [ ] Learn exported vs unexported names.
- [ ] Understand why capital letters matter in Go identifiers.
- [ ] Import our own package.

---

### Step 14 — First Git checkpoint

- [ ] Inspect `git diff`.
- [ ] Stage files.
- [ ] Commit the first working state.
- [ ] Write a meaningful commit message.

Suggested milestone:

```text
chore: initialize concord project
```

---

# PHASE 1 — Build a single-process key-value store

## Goal

Learn state, APIs, errors, tests, and encapsulation before adding a network.

---

### Step 15 — Design the Store API on paper

Before coding, decide what our store should support:

```text
Set(key, value)
Get(key)
Delete(key)
```

- [ ] Decide expected behavior for missing keys.
- [ ] Decide whether empty keys are allowed.
- [ ] Write the API before implementing it.

---

### Step 16 — Implement `Store`

Create something like:

```text
internal/store/store.go
```

- [ ] Define a `Store` struct.
- [ ] Give it an internal map.
- [ ] Write a constructor.
- [ ] Implement `Set`.
- [ ] Implement `Get`.
- [ ] Implement `Delete`.

Keep it single-threaded for now.

---

### Step 17 — Use the store from `main`

- [ ] Construct a store.
- [ ] Set a value.
- [ ] Read it.
- [ ] Delete it.
- [ ] Print results.
- [ ] Observe missing-key behavior.

---

### Step 18 — Learn Go tests

Create:

```text
internal/store/store_test.go
```

- [ ] Understand `testing.T`.
- [ ] Write the first test.
- [ ] Run `go test ./...`.
- [ ] Intentionally break the implementation.
- [ ] Watch the test fail.
- [ ] Fix it.

---

### Step 19 — Add table-driven tests

Test multiple cases using a slice of test cases.

- [ ] Learn anonymous structs.
- [ ] Learn loops with `range`.
- [ ] Learn why table-driven tests are common in Go.

---

### Step 20 — Define store errors intentionally

Possible example:

```text
ErrKeyNotFound
```

- [ ] Understand sentinel errors.
- [ ] Learn `errors.Is`.
- [ ] Improve tests around errors.

---

# PHASE 2 — Command parsing and a tiny protocol

## Goal

Turn human commands into structured operations.

---

### Step 21 — Define our first text protocol

Start with:

```text
SET key value
GET key
DELETE key
```

Responses might begin as:

```text
OK
VALUE alice
NOT_FOUND
ERROR ...
```

- [ ] Understand what a protocol is.
- [ ] Understand syntax vs semantics.
- [ ] Write examples before coding.

---

### Step 22 — Create a command representation

Create a `Command` type containing things such as:

```text
Type
Key
Value
```

- [ ] Consider string constants.
- [ ] Learn custom types.
- [ ] Learn why structured commands are better than passing raw strings everywhere.

---

### Step 23 — Implement command parsing

Create something like:

```text
internal/protocol/parser.go
```

- [ ] Split input.
- [ ] Validate argument counts.
- [ ] Reject unknown commands.
- [ ] Return useful errors.

---

### Step 24 — Test the parser thoroughly

Test:

- [ ] valid SET
- [ ] valid GET
- [ ] valid DELETE
- [ ] missing arguments
- [ ] unknown command
- [ ] blank input
- [ ] extra whitespace

---

### Step 25 — Build a command executor

Connect:

```text
Command -> Store -> Response
```

- [ ] Keep parsing separate from execution.
- [ ] Understand separation of concerns.
- [ ] Test executor behavior.

---

### Step 26 — Create a local REPL

Allow:

```text
$ go run .
concord> SET name alice
OK
concord> GET name
VALUE alice
```

- [ ] Learn `bufio.Scanner`.
- [ ] Learn standard input.
- [ ] Learn loops.
- [ ] Learn clean exit behavior.

This is still one process.

---

# PHASE 3 — Networking from first principles

## Goal

Turn Concord into a real server.

---

### Step 27 — Learn the networking mental model

Before coding, understand:

- [ ] IP address
- [ ] port
- [ ] TCP
- [ ] client
- [ ] server
- [ ] socket
- [ ] connection
- [ ] stream
- [ ] localhost
- [ ] why TCP is not "messages"

---

### Step 28 — Write the smallest TCP server

Create something like:

```text
internal/server/server.go
```

- [ ] Call `net.Listen`.
- [ ] Accept one connection.
- [ ] Print when a client connects.
- [ ] Close the connection.

Do not parse Concord commands yet.

---

### Step 29 — Connect using a basic TCP client tool

Possible tools:

```text
nc
telnet
```

- [ ] Connect to Concord.
- [ ] Observe the server accepting the connection.
- [ ] Understand what process owns the listening port.

---

### Step 30 — Read bytes from a connection

- [ ] Read incoming text.
- [ ] Print it on the server.
- [ ] Understand `io.Reader`.
- [ ] Understand byte slices.
- [ ] Understand EOF.

---

### Step 31 — Send a response

- [ ] Write bytes to the TCP connection.
- [ ] Observe them in the client.
- [ ] Understand request/response at a primitive level.

---

### Step 32 — Define message framing

Important lesson:

TCP gives us a **byte stream**, not command boundaries.

For the first version, use newline-delimited commands.

Example:

```text
SET name alice\n
```

- [ ] Understand framing.
- [ ] Understand why one `Read()` does not equal one request.
- [ ] Use buffered reading correctly.

---

### Step 33 — Connect TCP input to the protocol parser

Flow:

```text
TCP bytes
  -> line
  -> parser
  -> Command
```

- [ ] Parse network commands.
- [ ] Send parser errors back to the client.

---

### Step 34 — Connect commands to the store

Full flow:

```text
Client
 -> TCP
 -> parse
 -> execute
 -> Store
 -> response
 -> TCP
 -> Client
```

- [ ] SET works over TCP.
- [ ] GET works over TCP.
- [ ] DELETE works over TCP.

This is the first true Concord server.

---

### Step 35 — Build our own CLI client

Create eventually:

```text
cmd/concordctl/
```

Commands might become:

```bash
concordctl set name alice
concordctl get name
```

- [ ] Learn `os.Args`.
- [ ] Learn `net.Dial`.
- [ ] Understand why clients and servers are separate programs.

---

# PHASE 4 — Concurrency

## Goal

Support multiple clients and understand Go concurrency deeply.

---

### Step 36 — Demonstrate the single-client problem

- [ ] Connect one client and keep it open.
- [ ] Try connecting another.
- [ ] Observe the behavior.
- [ ] Understand blocking operations.

---

### Step 37 — Learn goroutines

- [ ] Learn what `go f()` means.
- [ ] Compare goroutines to OS threads conceptually.
- [ ] Handle each client connection in a goroutine.

---

### Step 38 — Trigger a race condition intentionally

Multiple clients now access the same Go map.

- [ ] Run with `go test -race` where appropriate.
- [ ] Understand what a data race is.
- [ ] Understand why maps are not safe for concurrent writes.

---

### Step 39 — Protect the store with a mutex

- [ ] Learn `sync.Mutex`.
- [ ] Learn lock/unlock.
- [ ] Use `defer`.
- [ ] Understand critical sections.
- [ ] Re-run race detection.

---

### Step 40 — Upgrade to `RWMutex`

- [ ] Understand readers vs writers.
- [ ] Use `RLock` for GET.
- [ ] Use `Lock` for SET/DELETE.
- [ ] Discuss when RWMutex helps and when it may not.

---

### Step 41 — Learn channels separately

Do not force channels into the store if a mutex is simpler.

- [ ] Build a tiny channel exercise.
- [ ] Learn send and receive.
- [ ] Learn buffered vs unbuffered channels.
- [ ] Learn `select`.
- [ ] Understand "share memory by communicating" without treating it as a law.

---

### Step 42 — Add connection lifecycle handling

- [ ] Handle disconnects.
- [ ] Close resources.
- [ ] Avoid leaking goroutines.
- [ ] Log connection errors sensibly.

---

### Step 43 — Add graceful shutdown

- [ ] Learn Unix signals at a beginner level.
- [ ] Catch SIGINT/SIGTERM.
- [ ] Stop accepting new connections.
- [ ] Close the listener.
- [ ] Wait for active goroutines where appropriate.
- [ ] Understand `sync.WaitGroup`.

---

### Step 44 — Introduce `context.Context`

- [ ] Understand cancellation.
- [ ] Understand deadlines at a high level.
- [ ] Pass context through server lifecycle where useful.
- [ ] Do not use context as a random key/value bag.

---

# PHASE 5 — Persistence

## Goal

Make data survive process crashes and restarts.

---

### Step 45 — Demonstrate data loss

- [ ] SET several keys.
- [ ] Kill Concord.
- [ ] Restart it.
- [ ] Observe that memory is gone.
- [ ] Define persistence.

---

### Step 46 — Learn basic file I/O

- [ ] Open a file.
- [ ] Write bytes.
- [ ] Read bytes.
- [ ] Close it.
- [ ] Understand file descriptors conceptually.

---

### Step 47 — Design a simple append-only log

Each mutation can be recorded as something like:

```text
SET name alice
DELETE age
```

- [ ] Understand append-only storage.
- [ ] Understand replay.
- [ ] Explain why GET does not need to be logged.

---

### Step 48 — Implement the first write-ahead log

Create something like:

```text
internal/wal/
```

- [ ] Append mutations.
- [ ] Keep the implementation intentionally simple.
- [ ] Test writes.

---

### Step 49 — Replay the WAL at startup

Startup becomes:

```text
open WAL
 -> replay commands
 -> rebuild in-memory state
 -> start server
```

- [ ] Restart Concord.
- [ ] Verify values survive.

---

### Step 50 — Understand crash consistency

Study scenarios like:

```text
memory updated
process crashes
disk write never happened
```

versus:

```text
disk write happened
memory update not completed
```

- [ ] Decide correct ordering.
- [ ] Understand why it is called a write-ahead log.

---

### Step 51 — Learn buffering and flushing

- [ ] OS page cache concept.
- [ ] buffered writer concept.
- [ ] `Flush`.
- [ ] `fsync` / `File.Sync`.
- [ ] durability vs performance trade-off.

---

### Step 52 — Give log records an explicit format

Move beyond arbitrary text.

Potential fields:

```text
length
operation
key
value
checksum
```

Do not over-engineer yet.

- [ ] Design before coding.

---

### Step 53 — Learn binary encoding basics

- [ ] bytes vs text
- [ ] integers as bytes
- [ ] endianness
- [ ] fixed-width numbers
- [ ] variable-length content
- [ ] encode/decode round trips

---

### Step 54 — Add checksums

- [ ] Understand accidental corruption.
- [ ] Learn CRC at a high level.
- [ ] Store a checksum per record.
- [ ] Reject corrupted records during replay.

---

### Step 55 — Handle partial final writes

Simulate a crash halfway through a record.

- [ ] Detect an incomplete tail record.
- [ ] Recover previously valid records.
- [ ] Avoid treating garbage as valid state.

---

### Step 56 — Add persistence tests

- [ ] temporary directories
- [ ] restart tests
- [ ] corrupt-record test
- [ ] partial-record test
- [ ] durability expectations

---

# PHASE 6 — Refactor into a clean single-node architecture

## Goal

Prepare the codebase for distributed behavior without prematurely implementing Raft.

---

### Step 57 — Separate major responsibilities

We should now be able to identify:

```text
server
protocol
store
wal
configuration
```

- [ ] Refactor only if responsibilities are currently mixed.
- [ ] Avoid abstraction for abstraction's sake.

---

### Step 58 — Create a real Concord node abstraction

A Node may own:

```text
ID
address
store
WAL
server
lifecycle state
```

- [ ] Understand ownership.
- [ ] Understand dependency injection without frameworks.

---

### Step 59 — Add configuration

Potential config:

```text
node ID
client address
data directory
```

- [ ] CLI flags.
- [ ] defaults.
- [ ] validation.
- [ ] understand configuration precedence.

---

### Step 60 — Improve logging

- [ ] Distinguish logs from normal client responses.
- [ ] Include node ID.
- [ ] Include useful context.
- [ ] Avoid noisy meaningless logs.

---

### Step 61 — Build a repeatable single-node integration test

Start a real server, connect a client, execute commands, restart it, verify persistence.

- [ ] Learn integration tests vs unit tests.

---

# PHASE 7 — Multi-node networking without consensus

## Goal

Make Concord nodes talk to other Concord nodes before Raft.

---

### Step 62 — Introduce peer configuration

Example concept:

```text
node1 -> node2, node3
node2 -> node1, node3
node3 -> node1, node2
```

- [ ] Understand node identity vs network address.
- [ ] Validate duplicate IDs.

---

### Step 63 — Separate client traffic from peer traffic

Potentially use different ports:

```text
client port
raft/peer port
```

- [ ] Understand why internal and external protocols differ.

---

### Step 64 — Build the smallest node-to-node RPC

Node A asks Node B:

```text
PING
```

Node B replies:

```text
PONG
```

- [ ] Implement manually.
- [ ] Understand RPC as "request another process to perform an operation."

---

### Step 65 — Add request IDs

- [ ] Understand correlation IDs.
- [ ] Match responses to requests.
- [ ] Discuss concurrency implications.

---

### Step 66 — Add peer connection timeouts

- [ ] connection timeout
- [ ] read timeout
- [ ] write timeout
- [ ] understand why distributed calls cannot wait forever

---

### Step 67 — Simulate a dead peer

- [ ] Shut one node down.
- [ ] Attempt RPC.
- [ ] Observe timeout/error.
- [ ] Treat network failures as normal events.

---

### Step 68 — Build a three-node local cluster

Run:

```text
node1
node2
node3
```

Each with unique:

- [ ] node ID
- [ ] client port
- [ ] peer port
- [ ] data directory

No automatic replication yet.

---

# PHASE 8 — Distributed systems foundations before Raft

## Goal

Understand the problem Raft solves before implementing Raft.

---

### Step 69 — Study failure models

Understand:

- [ ] process crash
- [ ] machine crash
- [ ] lost message
- [ ] delayed message
- [ ] duplicated message
- [ ] reordered message
- [ ] network partition
- [ ] disk failure
- [ ] Byzantine failure at a high level

Concord/Raft assumes non-Byzantine nodes.

---

### Step 70 — Learn the Two Generals intuition

- [ ] Understand why perfect agreement over unreliable communication is hard.
- [ ] Do not confuse it with Raft itself.

---

### Step 71 — Learn quorum / majority math

For cluster sizes:

```text
1
3
5
7
```

Calculate:

- [ ] majority
- [ ] tolerated failures
- [ ] why even-numbered clusters usually do not buy an extra failure

---

### Step 72 — Understand split brain

- [ ] Define split brain.
- [ ] Explain why two simultaneous leaders can corrupt consistency.
- [ ] Understand why majority ownership matters.

---

### Step 73 — Understand consensus

Consensus should answer roughly:

> Which ordered sequence of commands does the cluster agree happened?

- [ ] Differentiate replication from consensus.
- [ ] Differentiate leader election from full consensus.

---

### Step 74 — Learn Raft's high-level components

Understand:

- [ ] follower
- [ ] candidate
- [ ] leader
- [ ] term
- [ ] election timeout
- [ ] heartbeat
- [ ] RequestVote
- [ ] AppendEntries
- [ ] replicated log
- [ ] commit index
- [ ] applied index

No implementation yet.

---

### Step 75 — Walk through a Raft election by hand

For three nodes:

```text
A
B
C
```

Simulate:

- [ ] startup
- [ ] election timeout
- [ ] candidate transition
- [ ] vote requests
- [ ] majority
- [ ] leader
- [ ] heartbeats

---

### Step 76 — Walk through leader failure by hand

- [ ] A is leader.
- [ ] A dies.
- [ ] Followers stop receiving heartbeats.
- [ ] New election begins.
- [ ] A new leader emerges.

---

# PHASE 9 — Raft state machine and leader election

## Goal

Implement elections without log replication first.

---

### Step 77 — Create the Raft package

Eventually:

```text
internal/raft/
```

Define only the minimum foundational types.

- [ ] NodeState / Role
- [ ] Term
- [ ] Node ID

---

### Step 78 — Implement Raft role transitions

States:

```text
Follower
Candidate
Leader
```

- [ ] explicit transitions
- [ ] tests for transitions
- [ ] avoid networking initially

---

### Step 79 — Implement randomized election timeouts

- [ ] Understand why identical timeouts cause repeated split votes.
- [ ] Learn timers in Go.
- [ ] Randomize within a sensible interval.

---

### Step 80 — Candidate starts an election

On timeout:

- [ ] increment term
- [ ] vote for self
- [ ] transition to candidate
- [ ] reset election timer

Initially test state only.

---

### Step 81 — Define `RequestVote` messages

Fields will eventually include things such as:

```text
term
candidateID
lastLogIndex
lastLogTerm
```

For the first election implementation, introduce fields carefully.

---

### Step 82 — Implement vote-granting rules

At minimum:

- [ ] reject stale terms
- [ ] update on newer terms
- [ ] only vote once per term
- [ ] understand persistent vote state

---

### Step 83 — Persist current term and voted-for

This is essential Raft durable state.

- [ ] Decide storage format.
- [ ] Persist safely.
- [ ] Reload after restart.
- [ ] Test crash/restart behavior.

---

### Step 84 — Send RequestVote RPCs to peers

- [ ] issue peer RPCs
- [ ] handle timeouts
- [ ] collect responses
- [ ] do not block forever

---

### Step 85 — Count votes safely

- [ ] self vote
- [ ] peer votes
- [ ] majority detection
- [ ] concurrent response handling
- [ ] ignore stale responses

---

### Step 86 — Become leader after majority

- [ ] transition Candidate -> Leader.
- [ ] cancel/reset election timer behavior.
- [ ] log the leadership change.

---

### Step 87 — Define AppendEntries heartbeat messages

Initially entries can be empty.

- [ ] term
- [ ] leader ID
- [ ] other fields introduced intentionally

---

### Step 88 — Leader sends periodic heartbeats

- [ ] heartbeat ticker
- [ ] send to all peers
- [ ] handle failures

---

### Step 89 — Followers reset election timers on valid heartbeat

- [ ] prevent unnecessary elections.
- [ ] reject stale leaders.

---

### Step 90 — Step down when observing a newer term

Any leader/candidate seeing a newer term should become follower.

- [ ] RequestVote path
- [ ] AppendEntries path
- [ ] response path
- [ ] tests

---

### Step 91 — Test stable leader election with three nodes

- [ ] exactly one leader in stable conditions
- [ ] two followers
- [ ] wait several election periods
- [ ] leadership remains stable

---

### Step 92 — Kill the leader and observe re-election

- [ ] detect leader loss
- [ ] elect replacement
- [ ] restart old leader
- [ ] old leader rejoins as follower

This should feel like a major milestone.

---

# PHASE 10 — Raft replicated log

## Goal

Agree on an ordered sequence of commands.

---

### Step 93 — Define a Raft log entry

Potential fields:

```text
Index
Term
Command
```

- [ ] understand why term is attached to each entry.

---

### Step 94 — Store the Raft log in memory first

- [ ] append entry
- [ ] access by index
- [ ] inspect last index/term
- [ ] tests

---

### Step 95 — Update RequestVote with log freshness checks

Raft's voting rules include comparing candidate logs.

- [ ] understand "up-to-date log."
- [ ] implement lastLogTerm/lastLogIndex comparison.

---

### Step 96 — Expand AppendEntries fields

Introduce:

```text
prevLogIndex
prevLogTerm
entries
leaderCommit
```

- [ ] explain each field individually.

---

### Step 97 — Implement follower consistency checks

Follower validates:

```text
Does my log match prevLogIndex + prevLogTerm?
```

- [ ] reject inconsistent AppendEntries.
- [ ] accept matching prefixes.

---

### Step 98 — Implement conflict resolution

When logs disagree:

- [ ] identify conflicting entries.
- [ ] truncate unsafe suffix.
- [ ] append leader entries.

Understand why the leader's committed history wins.

---

### Step 99 — Leader tracks `nextIndex`

For every follower:

```text
nextIndex[peer]
```

- [ ] understand its purpose.
- [ ] initialize it when becoming leader.

---

### Step 100 — Leader tracks `matchIndex`

For every follower:

```text
matchIndex[peer]
```

- [ ] understand how it differs from nextIndex.

---

### Step 101 — Retry replication after rejection

- [ ] decrement/backtrack nextIndex.
- [ ] retry until logs align.
- [ ] optimize only later.

---

### Step 102 — Determine majority replication

For a log index:

- [ ] calculate how many nodes have replicated it.
- [ ] understand quorum.
- [ ] distinguish "replicated" from "committed."

---

### Step 103 — Advance `commitIndex`

Implement Raft's commit rules carefully.

- [ ] understand why leaders are conservative about entries from older terms.
- [ ] add targeted tests.

---

### Step 104 — Track `lastApplied`

Maintain:

```text
commitIndex
lastApplied
```

- [ ] explain why they are distinct.
- [ ] apply committed entries in order.

---

### Step 105 — Introduce the replicated state machine

Now:

```text
Raft log entry
 -> committed
 -> apply command
 -> key-value Store
```

- [ ] commands are deterministic.
- [ ] every healthy node reaches the same state.

---

# PHASE 11 — Route client writes through Raft

## Goal

Turn the Raft machinery back into a useful database.

---

### Step 106 — Prevent followers from directly accepting writes

When a client sends SET/DELETE to a follower:

- [ ] return "not leader" initially.
- [ ] include known leader information if available.

---

### Step 107 — Leader accepts a write as a proposed log entry

Flow:

```text
client SET
 -> leader
 -> append uncommitted Raft entry
 -> replicate
 -> majority
 -> commit
 -> apply
 -> reply success
```

- [ ] understand why response must not come too early.

---

### Step 108 — Wait for commit safely

- [ ] connect client request to eventual commit result.
- [ ] handle leadership loss while waiting.
- [ ] handle timeout.
- [ ] avoid goroutine leaks.

---

### Step 109 — Replicate DELETE

- [ ] model deletion as a deterministic state-machine command.
- [ ] test replication.

---

### Step 110 — Decide how GET works

Discuss options:

1. read only from leader
2. stale follower reads
3. linearizable reads

Start with the simplest safe policy.

- [ ] document the consistency guarantee.

---

### Step 111 — Verify identical state across all three nodes

- [ ] write several keys through leader.
- [ ] inspect each node.
- [ ] verify same committed state.

---

### Step 112 — Write during a follower failure

- [ ] stop one follower.
- [ ] keep leader + one follower alive.
- [ ] writes should still commit in a 3-node cluster.
- [ ] understand why.

---

### Step 113 — Lose quorum

- [ ] leave only one node alive in a three-node cluster.
- [ ] attempt a write.
- [ ] write must not commit.
- [ ] understand availability vs consistency.

---

### Step 114 — Recover a lagging follower

- [ ] follower goes offline.
- [ ] leader commits entries.
- [ ] follower returns.
- [ ] replication catches it up.

---

# PHASE 12 — Persistent Raft

## Goal

Make consensus survive restarts safely.

---

### Step 115 — Revisit all Raft persistent state

Raft requires persistence of important state such as:

- [ ] current term
- [ ] votedFor
- [ ] log entries

Review what must be durable before responding to RPCs.

---

### Step 116 — Design the persistent Raft log

Decide how the earlier WAL evolves or integrates.

- [ ] avoid having two contradictory sources of truth.
- [ ] define clear ownership.

---

### Step 117 — Persist log entries before acknowledging replication

- [ ] understand crash scenarios.
- [ ] enforce correct durability ordering.

---

### Step 118 — Recover Raft log after restart

- [ ] load term
- [ ] load vote
- [ ] load entries
- [ ] reconstruct indexes

---

### Step 119 — Rebuild state machine from committed data

Careful question:

How do we know which recovered entries were committed?

- [ ] understand Raft recovery implications.
- [ ] establish safe application behavior.

---

### Step 120 — Crash/restart one follower repeatedly

- [ ] confirm it rejoins.
- [ ] confirm it catches up.
- [ ] confirm no committed data is lost.

---

### Step 121 — Crash/restart the leader

- [ ] new leader election.
- [ ] old leader returns.
- [ ] logs converge.
- [ ] committed values remain.

---

### Step 122 — Crash all nodes and restart the cluster

- [ ] verify durable history.
- [ ] observe election after restart.
- [ ] verify state recovery.

This is another major milestone.

---

# PHASE 13 — Snapshots and log compaction

## Goal

Prevent the Raft log from growing forever.

---

### Step 123 — Understand why logs cannot grow forever

- [ ] disk usage
- [ ] startup replay time
- [ ] follower catch-up cost

---

### Step 124 — Define a state-machine snapshot

A snapshot represents key/value state at a specific Raft index/term.

Potential metadata:

```text
lastIncludedIndex
lastIncludedTerm
state
```

---

### Step 125 — Serialize store state

- [ ] choose an educational format.
- [ ] write snapshot.
- [ ] load snapshot.
- [ ] verify round trip.

---

### Step 126 — Write snapshots atomically

Learn the common pattern:

```text
write temp file
fsync
rename
```

- [ ] understand atomic rename assumptions.
- [ ] discuss directory fsync where relevant.

---

### Step 127 — Compact old log entries

Once safely represented by a snapshot:

- [ ] remove obsolete prefix.
- [ ] keep index math correct.

---

### Step 128 — Define InstallSnapshot RPC

- [ ] understand why a very far-behind follower cannot always catch up from retained log entries.
- [ ] design message fields.

---

### Step 129 — Send snapshot to lagging follower

- [ ] detect follower is behind compacted prefix.
- [ ] transfer snapshot.
- [ ] follower installs it.
- [ ] resume normal AppendEntries afterward.

---

### Step 130 — Test follower recovery via snapshot

- [ ] stop follower.
- [ ] generate many writes.
- [ ] compact.
- [ ] restart follower.
- [ ] observe snapshot installation.
- [ ] confirm final state.

---

# PHASE 14 — Cluster membership

## Goal

Understand why changing the set of voters is difficult.

---

### Step 131 — Study membership-change safety

- [ ] understand why simply editing a config file can create two majorities.
- [ ] learn joint consensus at a conceptual level.

---

### Step 132 — Decide educational scope

Choose whether Concord will implement:

- [ ] full Raft joint consensus
- [ ] a narrower safe membership mechanism

Document the decision.

---

### Step 133 — Implement adding a node safely

Only after the chosen membership design is understood.

- [ ] learner/non-voter stage if used
- [ ] catch up
- [ ] membership transition
- [ ] persist configuration

---

### Step 134 — Implement removing a node safely

- [ ] quorum implications
- [ ] leader removal case
- [ ] persistence

---

### Step 135 — Membership integration tests

- [ ] add node
- [ ] remove node
- [ ] restart cluster
- [ ] membership remains correct

---

# PHASE 15 — Client protocol improvements

## Goal

Make Concord pleasant and less fragile to use.

---

### Step 136 — Version the client protocol

- [ ] understand protocol evolution.
- [ ] add protocol version concept if justified.

---

### Step 137 — Structured request/response messages

Move beyond ad-hoc strings where useful.

Possible choices:

- JSON
- length-prefixed binary messages
- custom encoding

Compare trade-offs before choosing.

---

### Step 138 — Length-prefixed framing

If adopted:

```text
[message length][message bytes]
```

- [ ] partial reads
- [ ] exact reads
- [ ] maximum message size
- [ ] malformed length handling

---

### Step 139 — Add request IDs to client protocol

- [ ] correlate replies.
- [ ] enable future multiplexing concepts.

---

### Step 140 — Add proper error codes

Examples:

```text
NOT_LEADER
KEY_NOT_FOUND
INVALID_COMMAND
NO_QUORUM
TIMEOUT
INTERNAL_ERROR
```

- [ ] distinguish machine-readable code from human-readable message.

---

### Step 141 — Improve `concordctl`

Potential commands:

```bash
concordctl set foo bar
concordctl get foo
concordctl delete foo
concordctl status
concordctl members
```

- [ ] readable output
- [ ] non-zero exit codes for failures

---

# PHASE 16 — Observability

## Goal

Make invisible distributed behavior visible.

---

### Step 142 — Structured logs

Include fields such as:

```text
node_id
term
role
peer
event
```

Important events:

- [ ] election started
- [ ] vote granted
- [ ] became leader
- [ ] stepped down
- [ ] append rejected
- [ ] entry committed
- [ ] snapshot installed

---

### Step 143 — Add an HTTP health endpoint

Example:

```text
GET /health
```

- [ ] learn basic Go HTTP server.
- [ ] return node health.

---

### Step 144 — Add a status endpoint

Possible fields:

```json
{
  "node_id": "node1",
  "role": "leader",
  "term": 8,
  "commit_index": 142
}
```

- [ ] this endpoint is for observability, not consensus.

---

### Step 145 — Add metrics

Potential metrics:

```text
raft_term
raft_commit_index
raft_elections_total
raft_leader_changes_total
raft_rpc_failures_total
store_keys
```

- [ ] counters
- [ ] gauges
- [ ] latency histograms conceptually

---

### Step 146 — Measure RPC latency

- [ ] record peer RPC timings.
- [ ] observe effect of induced delays.

---

### Step 147 — Expose peer replication progress

For a leader, show:

```text
peer
nextIndex
matchIndex
```

This will be extremely useful for the dashboard.

---

# PHASE 17 — Failure testing and chaos

## Goal

Develop distributed-systems intuition by breaking Concord.

---

### Step 148 — Build a reproducible test cluster script

Eventually create scripts to:

- [ ] start 3 nodes
- [ ] stop all nodes
- [ ] clear test data
- [ ] show logs

We should first understand the manual commands.

---

### Step 149 — Add automated three-node integration tests

- [ ] start processes/nodes
- [ ] wait for leader
- [ ] write data
- [ ] verify replication
- [ ] shut down cleanly

---

### Step 150 — Test leader crash during write

Try failures at different moments:

```text
before local append
after local append
after one follower receives it
after majority replication
before client response
```

- [ ] reason about expected safety each time.

---

### Step 151 — Add artificial network delay

- [ ] inject configurable RPC delay.
- [ ] observe elections.
- [ ] observe replication lag.

---

### Step 152 — Add artificial packet/RPC drops

At the application transport layer:

- [ ] randomly drop selected RPCs.
- [ ] verify Raft remains safe.
- [ ] distinguish safety from liveness.

---

### Step 153 — Simulate a partition

Example:

```text
A | B C
```

- [ ] minority cannot commit.
- [ ] majority elects/keeps leader.
- [ ] heal partition.
- [ ] logs converge.

---

### Step 154 — Simulate a 1-1-1 partition

- [ ] no majority anywhere.
- [ ] no writes commit.
- [ ] observe repeated elections.

---

### Step 155 — Test slow follower behavior

- [ ] leader continues with majority.
- [ ] slow follower catches up later.
- [ ] understand backpressure concerns.

---

### Step 156 — Race detector and stress runs

- [ ] `go test -race ./...`
- [ ] repeated test loops
- [ ] identify flaky tests
- [ ] never accept unexplained flakes in consensus code

---

# PHASE 18 — Performance fundamentals

## Goal

Learn measurement without turning Concord into a premature optimization project.

---

### Step 157 — Create a simple benchmark client

Measure:

- [ ] requests/sec
- [ ] latency
- [ ] payload sizes
- [ ] concurrent clients

---

### Step 158 — Learn Go benchmarks

- [ ] `BenchmarkXxx`
- [ ] `go test -bench`
- [ ] allocations
- [ ] benchmark pitfalls

---

### Step 159 — Profile CPU

- [ ] learn `pprof` conceptually.
- [ ] find actual hot paths.
- [ ] avoid guessing.

---

### Step 160 — Profile memory

- [ ] heap profile
- [ ] allocations
- [ ] goroutine count
- [ ] look for leaks

---

### Step 161 — Evaluate WAL batching

Compare:

```text
fsync every write
batch multiple writes
```

- [ ] durability trade-off
- [ ] latency trade-off
- [ ] throughput trade-off

Do not change defaults without documenting semantics.

---

# PHASE 19 — Security and defensive engineering

## Goal

Learn how servers protect themselves from malformed or abusive input.

---

### Step 162 — Add request size limits

- [ ] reject gigantic keys/values/messages.
- [ ] avoid accidental memory exhaustion.

---

### Step 163 — Add connection limits

- [ ] understand resource exhaustion.
- [ ] limit concurrent clients if appropriate.

---

### Step 164 — Harden protocol parsing

Test:

- [ ] malformed lengths
- [ ] invalid UTF-8 if relevant
- [ ] unexpected message types
- [ ] truncated messages
- [ ] duplicate fields
- [ ] impossible indexes

---

### Step 165 — Add fuzz tests

Use Go fuzzing for parsers/decoders.

- [ ] learn fuzzing mental model.
- [ ] ensure malformed bytes do not panic the server.

---

### Step 166 — Consider peer authentication

For the educational version, study:

- [ ] why internal cluster RPCs should eventually be authenticated.
- [ ] TLS.
- [ ] mutual TLS.

Implementation can be optional depending on scope.

---

# PHASE 20 — Docker and reproducible deployment

## Goal

Run Concord consistently across machines.

---

### Step 167 — Learn containers before writing Dockerfiles

Understand:

- [ ] image
- [ ] container
- [ ] filesystem layer
- [ ] port mapping
- [ ] volume
- [ ] process isolation
- [ ] container vs VM

---

### Step 168 — Dockerize one Concord node

- [ ] write Dockerfile manually.
- [ ] build image.
- [ ] run container.
- [ ] connect with concordctl.

---

### Step 169 — Persist data through a Docker volume

- [ ] restart container.
- [ ] verify data survives.

---

### Step 170 — Build a three-node Docker Compose cluster

- [ ] separate node identities.
- [ ] private container network.
- [ ] persistent volumes.
- [ ] expose useful client/API ports.

---

### Step 171 — Kill one container

- [ ] observe leader failover.
- [ ] write through surviving majority.
- [ ] restart node.
- [ ] observe catch-up.

---

# PHASE 21 — Dashboard / frontend

## Goal

Create a recruiter-friendly visual demonstration without hiding the distributed system.

---

### Step 172 — Design the dashboard information architecture

Possible display:

```text
Cluster
 ├─ node1 LEADER   term 8   commit 150
 ├─ node2 FOLLOWER term 8   commit 150
 └─ node3 FOLLOWER term 8   commit 149
```

Additional panels:

- recent elections
- replication lag
- key count
- RPC health
- logs/events

---

### Step 173 — Create a minimal static frontend

Potential stack:

- plain HTML
- CSS
- small JavaScript

Avoid adding React unless there is a real reason.

---

### Step 174 — Poll the status API

- [ ] fetch node status.
- [ ] display role/term.
- [ ] handle node unavailable state.

---

### Step 175 — Visualize cluster topology

Show leader/follower relationships.

- [ ] distinguish live data from decorative graphics.

---

### Step 176 — Add key-value operations to the dashboard

Optional:

- [ ] SET
- [ ] GET
- [ ] DELETE

The CLI should remain supported.

---

### Step 177 — Add a failure-demo control only for local/demo mode

Potential controls:

```text
pause node
inject latency
drop RPCs
```

Never expose unsafe debugging controls in public production mode.

---

# PHASE 22 — Public demo deployment

## Goal

Let recruiters interact with a safe demo.

---

### Step 178 — Define deployment modes

Likely:

```text
local real cluster
public demo/simulation mode
```

Because free hosting platforms may have networking, persistence, sleep, or process limitations.

Document what is real and what is simulated.

---

### Step 179 — Deploy a public recruiter-friendly instance

Potentially Render or another service depending on current free-tier constraints.

At this step, check current hosting capabilities rather than relying on old assumptions.

- [ ] backend reachable
- [ ] dashboard reachable
- [ ] health endpoint
- [ ] safe resource limits

---

### Step 180 — Prevent abuse in public demo

- [ ] request limits
- [ ] value size limits
- [ ] rate limits if needed
- [ ] no arbitrary filesystem access
- [ ] no unrestricted chaos controls
- [ ] automatic reset if useful

---

# PHASE 23 — Real three-machine / VM demo

## Goal

Demonstrate Concord as an actual distributed system across separate environments.

---

### Step 181 — Prepare three environments

Could use:

- three VMs
- old laptop + local VMs
- multiple physical machines
- cloud VMs if available

Each node must have:

- [ ] unique IP/address
- [ ] unique node ID
- [ ] reachable peer ports
- [ ] persistent data directory

---

### Step 182 — Run Concord across three nodes

- [ ] establish peer connectivity.
- [ ] elect leader.
- [ ] replicate writes.
- [ ] confirm state.

---

### Step 183 — Record a leader-failure demo

Demo script:

1. show current leader
2. write data
3. kill leader
4. observe election
5. write again
6. restart old leader
7. observe catch-up
8. verify all nodes agree

---

### Step 184 — Record a partition demo

If the environment allows firewall/network rules:

- [ ] isolate a minority.
- [ ] demonstrate no minority writes.
- [ ] continue on majority.
- [ ] heal network.
- [ ] show convergence.

---

# PHASE 24 — Documentation for interviews

## Goal

Be able to explain Concord clearly, not merely show code.

---

### Step 185 — Write a strong README

Include:

- [ ] what Concord is
- [ ] why it exists
- [ ] architecture
- [ ] features
- [ ] quick start
- [ ] cluster demo
- [ ] consistency model
- [ ] limitations
- [ ] learning goals

---

### Step 186 — Draw the architecture diagram

Include:

- client
- leader
- followers
- Raft log
- state machine
- WAL
- snapshots
- metrics/dashboard

---

### Step 187 — Document the write path

Explain:

```text
client
 -> leader
 -> local log
 -> followers
 -> quorum
 -> commit
 -> state machine
 -> response
```

---

### Step 188 — Document leader election

Explain in our own words:

- term
- timeout
- vote
- quorum
- heartbeat
- failover

---

### Step 189 — Document failure guarantees

Answer:

- What happens if one node dies?
- What happens if two of three nodes die?
- What happens during a partition?
- Can stale reads happen?
- When is a write acknowledged?
- What survives a total restart?

---

### Step 190 — Document engineering trade-offs

Examples:

- TCP vs HTTP for peer transport
- text vs binary protocol
- fsync policy
- leader-only reads
- mutex vs actor-style state
- custom Raft implementation vs library
- simple snapshot format
- deployment constraints

---

### Step 191 — Write "What I learned"

This is important for interviews.

Topics:

- Go
- goroutines
- synchronization
- TCP
- persistence
- crash safety
- quorum
- Raft
- testing distributed code
- observability
- deployment

---

# PHASE 25 — Interview hardening

## Goal

Be able to defend every major design decision.

---

### Step 192 — Explain Concord in 30 seconds

Practice a concise answer.

---

### Step 193 — Explain Concord in 3 minutes

Cover:

- problem
- architecture
- Raft
- persistence
- failure handling
- most difficult lesson

---

### Step 194 — Whiteboard Raft leader election

No code.

---

### Step 195 — Whiteboard a client write

No code.

---

### Step 196 — Explain a network partition

Use a 3-node example.

---

### Step 197 — Explain crash consistency

Describe WAL ordering and fsync.

---

### Step 198 — Explain one difficult bug we encountered

Keep a real debugging story during development.

---

### Step 199 — Compare Concord with etcd/Consul/Redis

Do not claim feature parity.

Explain what real systems contain that Concord intentionally does not.

---

### Step 200 — Final project audit

Verify:

- [ ] tests pass
- [ ] race detector passes
- [ ] README accurate
- [ ] demo instructions work
- [ ] no secrets in repository
- [ ] logs are understandable
- [ ] code is formatted
- [ ] dead code removed
- [ ] limitations documented
- [ ] architecture diagram current
- [ ] demo video recorded
- [ ] public demo works if still hosted

---

# 9. Optional advanced extensions

These are deliberately **outside the first 200 steps**.

Do not start them merely because they sound impressive.

Possible future directions:

## A. Linearizable reads

Learn:

- ReadIndex
- lease reads and their caveats
- quorum confirmation

## B. Client sessions / deduplication

Handle:

```text
client retries after timeout
```

without accidentally executing the same logical write twice.

Learn:

- request IDs
- idempotency
- exactly-once illusion

## C. Transactions

Start with very small atomic compare-and-set operations.

## D. Watch API

Clients subscribe to changes:

```text
WATCH foo
```

Learn:

- long-lived streams
- event indexes
- missed-event recovery
- backpressure

## E. TTL / leases

Learn:

- replicated time-related behavior
- why wall clocks in distributed systems are tricky

## F. Read-only followers

Explicitly support stale reads and document semantics.

## G. Joint consensus membership changes

If not already implemented fully.

## H. TLS / mTLS

Secure client and peer communication.

## I. Prometheus + Grafana

Use standard observability tooling.

## J. Jepsen-style testing ideas

Build a smaller custom linearizability/fault harness.

## K. Kubernetes deployment

Only after understanding the system without Kubernetes.

## L. gRPC transport comparison

Reimplement or prototype the transport using gRPC and compare it to our hand-built protocol.

---

# 10. Concepts we should understand by the end

This is a knowledge checklist, not an implementation checklist.

## Go

- [ ] packages
- [ ] modules
- [ ] variables
- [ ] structs
- [ ] interfaces
- [ ] methods
- [ ] pointers
- [ ] slices
- [ ] maps
- [ ] errors
- [ ] defer
- [ ] goroutines
- [ ] channels
- [ ] mutexes
- [ ] RWMutex
- [ ] WaitGroup
- [ ] context
- [ ] timers
- [ ] interfaces
- [ ] io.Reader / io.Writer
- [ ] net.Conn
- [ ] testing
- [ ] benchmarks
- [ ] fuzzing
- [ ] profiling
- [ ] race detector

## Networking

- [ ] IP
- [ ] TCP
- [ ] ports
- [ ] sockets
- [ ] client/server
- [ ] byte streams
- [ ] message framing
- [ ] timeouts
- [ ] connection failure
- [ ] serialization
- [ ] RPC
- [ ] partial reads/writes

## Operating systems / storage

- [ ] process
- [ ] thread
- [ ] file descriptor
- [ ] memory vs disk
- [ ] page cache
- [ ] buffering
- [ ] flush
- [ ] fsync
- [ ] atomic rename
- [ ] signals
- [ ] graceful shutdown

## Database/storage ideas

- [ ] in-memory state
- [ ] durability
- [ ] write-ahead log
- [ ] append-only log
- [ ] checksums
- [ ] replay
- [ ] snapshots
- [ ] log compaction
- [ ] crash consistency

## Distributed systems

- [ ] partial failure
- [ ] replication
- [ ] consensus
- [ ] quorum
- [ ] majority
- [ ] leader election
- [ ] terms
- [ ] heartbeats
- [ ] network partition
- [ ] split brain
- [ ] stale nodes
- [ ] safety
- [ ] liveness
- [ ] consistency
- [ ] availability
- [ ] replicated state machine
- [ ] deterministic commands
- [ ] log matching
- [ ] commit index
- [ ] failure recovery

## Engineering

- [ ] unit tests
- [ ] integration tests
- [ ] failure tests
- [ ] fuzz tests
- [ ] observability
- [ ] structured logs
- [ ] metrics
- [ ] profiling
- [ ] Docker
- [ ] deployment
- [ ] documentation
- [ ] reproducible demos

---

# 11. Milestones

These are the moments worth celebrating.

## Milestone 1 — First Go binary

Concord prints something and we understand every line.

## Milestone 2 — Local KV store

SET/GET/DELETE work in memory.

## Milestone 3 — Networked Concord

A separate TCP client talks to the server.

## Milestone 4 — Concurrent server

Multiple clients work safely.

## Milestone 5 — Durable single-node Concord

Data survives restart.

## Milestone 6 — Three nodes can communicate

Peer RPC works.

## Milestone 7 — Leader election works

Kill the leader and another node becomes leader.

## Milestone 8 — Replicated writes work

A leader write reaches all healthy nodes.

## Milestone 9 — Quorum behavior works

Minority cannot commit writes.

## Milestone 10 — Full restart works

Cluster crashes and recovers without losing committed data.

## Milestone 11 — Snapshot recovery works

A far-behind follower catches up from a snapshot.

## Milestone 12 — Chaos demo works

Partitions, delays, and node crashes behave as expected.

## Milestone 13 — Recruiter demo

Dashboard + documentation + recorded multi-node failure demo.

---

# 12. Progress Tracker

Update this section over time.

```text
Current step: 1
Current phase: Phase 0 — Environment, Go basics, and project foundations
Status: NOT STARTED

Last completed step: None

Current repository state:
- Repository may not yet exist.
- No assumptions should be made beyond this roadmap.

Current conceptual knowledge explicitly established:
- None yet. Start from first principles.

Next action:
- Complete Step 1 only.
```

---

# 13. Session handoff template

At the end of a coding session, add a note like this:

```text
SESSION HANDOFF

Date:
Completed through step:
Current step:
Files created:
Files modified:
Commands run:
Tests passing:
Known bugs:
Concepts learned:
Decisions made:
Things I am still confused about:
Exact next task:
```

This makes it much easier for another AI to continue reliably.

---

# 14. Git milestone suggestions

These are only examples.

```text
chore: initialize concord project
feat: add in-memory key value store
feat: add command parser
feat: add tcp server
feat: support concurrent clients
feat: add write ahead log
feat: recover state from wal
feat: add peer transport
feat: implement raft elections
feat: replicate raft log entries
feat: apply committed commands
feat: persist raft state
feat: add snapshots
feat: add cluster status api
feat: add concord dashboard
test: add partition failure scenarios
docs: document concord architecture
```

Do not commit blindly at every tiny edit.

A commit should represent a coherent state worth returning to.

---

# 15. Important project philosophy

## We are building understanding, not merely software.

A 200-step roadmap is not excessive if each step teaches something useful.

If one step turns out to contain three concepts I do not understand, split it into smaller steps.

If the code works but I cannot explain why, that step is **not complete**.

If an abstraction hides the thing I am trying to learn, avoid that abstraction until later.

If we discover our earlier implementation was naive, that is not failure.

That is the point of building the system incrementally.

---

# 16. Definition of "Step Complete"

A step is complete only when most of these are true:

1. The code runs, if the step contains code.
2. Relevant tests pass.
3. I can explain what we changed.
4. I can explain why we changed it.
5. I understand the new Go concepts introduced.
6. I understand the new systems concept introduced.
7. I can describe at least one way the implementation can fail.
8. I know what the next step builds on top of this one.

---

# 17. Instructions to the next AI

If you are an AI receiving this file:

1. Read the entire roadmap before continuing.
2. Find the Progress Tracker.
3. Respect the current step.
4. Do **not** assume prior knowledge beyond explicitly completed steps.
5. Teach first principles.
6. Work in small increments.
7. Do not generate the entire final implementation.
8. The human creates files and folders manually unless they explicitly ask you to do it.
9. Explain every new abstraction.
10. Keep the project runnable as often as possible.
11. Prefer tests and observable experiments.
12. When a distributed-systems behavior is subtle, walk through concrete node-by-node examples.
13. Never trade away understanding merely to move faster.
14. Update this roadmap/progress information when useful.
15. If the roadmap's design is wrong, explain the issue before changing direction.

The immediate goal is **not** to finish Concord.

The immediate goal is to complete the **next step while understanding it**.

---

# 18. Start here

The next session should begin with:

> **Step 1 — Define the project in our own words**

Do not create code yet.

First explain, from first principles:

1. What is a key-value store?
2. What does distributed mean?
3. What is replication?
4. Why might several replicas disagree?
5. Why do we eventually need consensus?
6. What exactly will Concord do when it is finished?

Then I should explain Concord back in my own words.

Only after that should we proceed to Step 2.

