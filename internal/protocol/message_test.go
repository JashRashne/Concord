package protocol

import (
	"encoding/json"
	"testing"
)

func TestMessageJSONRoundTrip(t *testing.T) {
	original := Message{
		Type: MessageTypePing,
		From: "node-1",
		Data: "hello",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	var decoded Message

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	if decoded.Type != original.Type {
		t.Errorf("expected type %q, got %q", original.Type, decoded.Type)
	}

	if decoded.From != original.From {
		t.Errorf("expected from %q, got %q", original.From, decoded.From)
	}

	if decoded.Data != original.Data {
		t.Errorf("expected data %q, got %q", original.Data, decoded.Data)
	}
}

func TestRequestVoteJSONRoundTrip(t *testing.T) {
	original := Message{
		Type: MessageTypeRequestVote,
		From: "node-2",
		RequestVote: &RequestVoteRequest{
			Term:         4,
			CandidateID:  "node-2",
			LastLogIndex: 0,
			LastLogTerm:  0,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Message

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.RequestVote == nil {
		t.Fatal("expected RequestVote payload")
	}

	if decoded.RequestVote.Term != 4 {
		t.Fatalf(
			"expected term 4, got %d",
			decoded.RequestVote.Term,
		)
	}

	if decoded.RequestVote.CandidateID != "node-2" {
		t.Fatalf(
			"expected node-2, got %s",
			decoded.RequestVote.CandidateID,
		)
	}
}

func TestAppendEntriesJSONRoundTrip(t *testing.T) {
	original := Message{
		Type: MessageTypeAppendEntries,
		From: "node-2",
		AppendEntries: &AppendEntriesRequest{
			Term:     7,
			LeaderID: "node-2",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Message

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.AppendEntries == nil {
		t.Fatal("expected AppendEntries payload")
	}

	if decoded.AppendEntries.Term != 7 {
		t.Fatalf(
			"expected term 7, got %d",
			decoded.AppendEntries.Term,
		)
	}

	if decoded.AppendEntries.LeaderID != "node-2" {
		t.Fatalf(
			"expected node-2, got %s",
			decoded.AppendEntries.LeaderID,
		)
	}
}
