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
