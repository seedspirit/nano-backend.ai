package common

import (
	"encoding/json"
	"errors"
	"strings"
)

// ErrInvalidImageRef indicates the string is not a valid container image reference.
var ErrInvalidImageRef = errors.New("invalid image reference")

// ImageRef is an opaque container image reference with registry, repository, and tag.
// The zero value is invalid; use ParseImageRef to construct.
type ImageRef struct {
	registry   string
	repository string
	tag        string
}

// ParseImageRef parses a Docker image reference string.
// Format: [registry/]repository[:tag]
// Defaults: registry → "docker.io", tag → "latest".
// Single-segment names (e.g. "nginx") are expanded to "library/nginx".
func ParseImageRef(s string) (ImageRef, error) {
	if s == "" {
		return ImageRef{}, ErrInvalidImageRef
	}

	// Split off tag.
	repo := s
	tag := "latest"
	if idx := strings.LastIndex(repo, ":"); idx != -1 {
		tag = repo[idx+1:]
		repo = repo[:idx]
	}

	if repo == "" || tag == "" {
		return ImageRef{}, ErrInvalidImageRef
	}

	registry, repository := splitRegistryRepo(repo)

	return ImageRef{
		registry:   registry,
		repository: repository,
		tag:        tag,
	}, nil
}

// splitRegistryRepo separates the registry from the repository path.
// A path segment is treated as a registry if it contains a dot or colon
// (e.g. "registry.example.com", "localhost:5000").
func splitRegistryRepo(repo string) (registry, repository string) {
	parts := strings.SplitN(repo, "/", 2)

	if len(parts) == 1 {
		// Single name like "nginx" → docker.io/library/nginx
		return "docker.io", "library/" + repo
	}

	first := parts[0]
	if strings.ContainsAny(first, ".:") {
		// Custom registry: "registry.example.com/org/app"
		return first, parts[1]
	}

	// Docker Hub user repo: "myuser/myapp"
	return "docker.io", repo
}

// Registry returns the registry host (e.g. "docker.io").
func (r ImageRef) Registry() string { return r.registry }

// Repository returns the repository path (e.g. "library/nginx").
func (r ImageRef) Repository() string { return r.repository }

// Tag returns the image tag (e.g. "latest").
func (r ImageRef) Tag() string { return r.tag }

// String returns the canonical form "registry/repository:tag".
func (r ImageRef) String() string {
	if r.IsZero() {
		return ""
	}
	return r.registry + "/" + r.repository + ":" + r.tag
}

// IsZero reports whether the ImageRef is the zero value (uninitialized).
func (r ImageRef) IsZero() bool {
	return r.registry == "" && r.repository == "" && r.tag == ""
}

// MarshalJSON implements json.Marshaler.
func (r ImageRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *ImageRef) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseImageRef(s)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}
