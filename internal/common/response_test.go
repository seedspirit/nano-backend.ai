package common

import (
	"encoding/json"
	"testing"
)

func TestNewApiResponseSetsAllFields(t *testing.T) {
	resp := NewApiResponse("healthy", "service is running", "proceed with requests")

	if resp.Status != "healthy" {
		t.Errorf("got status %q, want %q", resp.Status, "healthy")
	}
	if resp.Reason != "service is running" {
		t.Errorf("got reason %q, want %q", resp.Reason, "service is running")
	}
	if resp.NextActionHint != "proceed with requests" {
		t.Errorf("got next_action_hint %q, want %q", resp.NextActionHint, "proceed with requests")
	}
}

func TestOkResponseSetsStatusToOk(t *testing.T) {
	resp := OkResponse("done", "continue")

	if resp.Status != "ok" {
		t.Errorf("got status %q, want %q", resp.Status, "ok")
	}
}

func TestErrorResponseSetsStatusToError(t *testing.T) {
	resp := ErrorResponse("failed", "retry later")

	if resp.Status != "error" {
		t.Errorf("got status %q, want %q", resp.Status, "error")
	}
}

func TestApiResponseSerializesToJSON(t *testing.T) {
	resp := NewApiResponse("healthy", "all good", "send requests")

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if m["status"] != "healthy" {
		t.Errorf("got status %q, want %q", m["status"], "healthy")
	}
	if m["reason"] != "all good" {
		t.Errorf("got reason %q, want %q", m["reason"], "all good")
	}
	if m["next_action_hint"] != "send requests" {
		t.Errorf("got next_action_hint %q, want %q", m["next_action_hint"], "send requests")
	}
}

func TestApiResponseDeserializesFromJSON(t *testing.T) {
	input := `{"status":"ok","reason":"done","next_action_hint":"proceed"}`

	var resp ApiResponse
	if err := json.Unmarshal([]byte(input), &resp); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("got status %q, want %q", resp.Status, "ok")
	}
	if resp.Reason != "done" {
		t.Errorf("got reason %q, want %q", resp.Reason, "done")
	}
	if resp.NextActionHint != "proceed" {
		t.Errorf("got next_action_hint %q, want %q", resp.NextActionHint, "proceed")
	}
}
