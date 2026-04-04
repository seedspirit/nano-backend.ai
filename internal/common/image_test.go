package common

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestParseImageRefSimpleName(t *testing.T) {
	ref, err := ParseImageRef("nginx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Registry() != "docker.io" {
		t.Errorf("got registry %q, want %q", ref.Registry(), "docker.io")
	}
	if ref.Repository() != "library/nginx" {
		t.Errorf("got repo %q, want %q", ref.Repository(), "library/nginx")
	}
	if ref.Tag() != "latest" {
		t.Errorf("got tag %q, want %q", ref.Tag(), "latest")
	}
}

func TestParseImageRefWithTag(t *testing.T) {
	ref, err := ParseImageRef("nginx:1.25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Repository() != "library/nginx" {
		t.Errorf("got repo %q, want %q", ref.Repository(), "library/nginx")
	}
	if ref.Tag() != "1.25" {
		t.Errorf("got tag %q, want %q", ref.Tag(), "1.25")
	}
}

func TestParseImageRefUserRepo(t *testing.T) {
	ref, err := ParseImageRef("myuser/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Registry() != "docker.io" {
		t.Errorf("got registry %q, want %q", ref.Registry(), "docker.io")
	}
	if ref.Repository() != "myuser/myapp" {
		t.Errorf("got repo %q, want %q", ref.Repository(), "myuser/myapp")
	}
	if ref.Tag() != "latest" {
		t.Errorf("got tag %q, want %q", ref.Tag(), "latest")
	}
}

func TestParseImageRefUserRepoWithTag(t *testing.T) {
	ref, err := ParseImageRef("myuser/myapp:v2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Repository() != "myuser/myapp" {
		t.Errorf("got repo %q, want %q", ref.Repository(), "myuser/myapp")
	}
	if ref.Tag() != "v2" {
		t.Errorf("got tag %q, want %q", ref.Tag(), "v2")
	}
}

func TestParseImageRefCustomRegistry(t *testing.T) {
	ref, err := ParseImageRef("registry.example.com/myapp:v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Registry() != "registry.example.com" {
		t.Errorf("got registry %q, want %q", ref.Registry(), "registry.example.com")
	}
	if ref.Repository() != "myapp" {
		t.Errorf("got repo %q, want %q", ref.Repository(), "myapp")
	}
	if ref.Tag() != "v1" {
		t.Errorf("got tag %q, want %q", ref.Tag(), "v1")
	}
}

func TestParseImageRefCustomRegistryWithOrg(t *testing.T) {
	ref, err := ParseImageRef("registry.example.com/org/myapp:v1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.Registry() != "registry.example.com" {
		t.Errorf("got registry %q, want %q", ref.Registry(), "registry.example.com")
	}
	if ref.Repository() != "org/myapp" {
		t.Errorf("got repo %q, want %q", ref.Repository(), "org/myapp")
	}
	if ref.Tag() != "v1" {
		t.Errorf("got tag %q, want %q", ref.Tag(), "v1")
	}
}

func TestParseImageRefEmpty(t *testing.T) {
	_, err := ParseImageRef("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
	if !errors.Is(err, ErrInvalidImageRef) {
		t.Errorf("got %v, want ErrInvalidImageRef", err)
	}
}

func TestParseImageRefColonOnly(t *testing.T) {
	_, err := ParseImageRef(":tag")
	if err == nil {
		t.Fatal("expected error for colon-only string")
	}
	if !errors.Is(err, ErrInvalidImageRef) {
		t.Errorf("got %v, want ErrInvalidImageRef", err)
	}
}

func TestImageRefStringCanonical(t *testing.T) {
	ref, err := ParseImageRef("nginx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "docker.io/library/nginx:latest"
	if ref.String() != want {
		t.Errorf("got %q, want %q", ref.String(), want)
	}
}

func TestImageRefIsZeroValid(t *testing.T) {
	ref, _ := ParseImageRef("nginx")
	if ref.IsZero() {
		t.Error("expected non-zero ImageRef from ParseImageRef")
	}
}

func TestImageRefIsZeroDefault(t *testing.T) {
	var ref ImageRef
	if !ref.IsZero() {
		t.Error("expected zero-value ImageRef to be zero")
	}
}

func TestImageRefEquality(t *testing.T) {
	a, _ := ParseImageRef("nginx:latest")
	b, _ := ParseImageRef("nginx")

	if a != b {
		t.Errorf("expected %q == %q (both resolve to same canonical form)", a, b)
	}

	c, _ := ParseImageRef("nginx:1.25")
	if a == c {
		t.Errorf("expected %q != %q", a, c)
	}
}

func TestImageRefJSONRoundtrip(t *testing.T) {
	original, _ := ParseImageRef("registry.example.com/myapp:v1")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded ImageRef
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if decoded != original {
		t.Errorf("got %q, want %q", decoded, original)
	}
}

func TestImageRefJSONInvalid(t *testing.T) {
	var ref ImageRef
	err := json.Unmarshal([]byte(`""`), &ref)
	if err == nil {
		t.Fatal("expected error for empty string in JSON")
	}
}
