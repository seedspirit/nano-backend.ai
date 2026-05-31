package workload

import (
	"encoding/json"
	"testing"
)

func TestParseImageRefValid(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantRegistry   string
		wantRepository string
		wantTag        string
	}{
		{
			name:           "repository only defaults tag to latest",
			input:          "python",
			wantRegistry:   "",
			wantRepository: "python",
			wantTag:        "latest",
		},
		{
			name:           "repository with tag",
			input:          "python:3.11",
			wantRepository: "python",
			wantTag:        "3.11",
		},
		{
			name:           "namespaced repository is not treated as registry",
			input:          "library/python:3.11",
			wantRepository: "library/python",
			wantTag:        "3.11",
		},
		{
			name:           "registry with dot",
			input:          "ghcr.io/org/img:v1",
			wantRegistry:   "ghcr.io",
			wantRepository: "org/img",
			wantTag:        "v1",
		},
		{
			name:           "registry without tag defaults to latest",
			input:          "ghcr.io/org/img",
			wantRegistry:   "ghcr.io",
			wantRepository: "org/img",
			wantTag:        "latest",
		},
		{
			name:           "localhost registry with port",
			input:          "localhost:5000/img:v2",
			wantRegistry:   "localhost:5000",
			wantRepository: "img",
			wantTag:        "v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ParseImageRef(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ref.Registry() != tt.wantRegistry {
				t.Errorf("registry: got %q, want %q", ref.Registry(), tt.wantRegistry)
			}
			if ref.Repository() != tt.wantRepository {
				t.Errorf("repository: got %q, want %q", ref.Repository(), tt.wantRepository)
			}
			if ref.Tag() != tt.wantTag {
				t.Errorf("tag: got %q, want %q", ref.Tag(), tt.wantTag)
			}
		})
	}
}

func TestParseImageRefInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "whitespace only", input: "   "},
		{name: "empty repository with tag", input: ":v1"},
		{name: "empty tag", input: "python:"},
		{name: "registry only with empty repository", input: "ghcr.io/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseImageRef(tt.input); err == nil {
				t.Fatalf("expected error for input %q", tt.input)
			}
		})
	}
}

func TestImageRefStringCanonical(t *testing.T) {
	ref, err := ParseImageRef("ghcr.io/org/img")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := ref.String(), "ghcr.io/org/img:latest"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestImageRefJSONRoundtrip(t *testing.T) {
	original, err := ParseImageRef("ghcr.io/org/img:v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if got, want := string(data), `"ghcr.io/org/img:v1"`; got != want {
		t.Errorf("marshaled: got %s, want %s", got, want)
	}

	var decoded ImageRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded != original {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestImageRefJSONInvalid(t *testing.T) {
	var ref ImageRef
	if err := json.Unmarshal([]byte(`":v1"`), &ref); err == nil {
		t.Fatal("expected error for invalid image ref in JSON")
	}
}
