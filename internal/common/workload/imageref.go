// Package workload defines the minimal manager-agent execution contract: the
// Plan an agent prepares and starts, the Launcher port the manager drives, and
// the Ref that routes follow-up calls. It is transport- and substrate-agnostic:
// no Docker SDK types, no HTTP/REST types, no manager-only scheduling policy.
package workload

import (
	"strings"

	"github.com/seedspirit/nano-backend.ai/internal/common/encoding"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// ImageRef is an opaque container image reference of the form
// [registry/]repository[:tag]. The zero value is invalid; use ParseImageRef.
type ImageRef struct {
	registry   string
	repository string
	tag        string
}

const defaultTag = "latest"

// ParseImageRef parses [registry/]repository[:tag]. The first path component is
// a registry only when it looks like a host (has '.' or ':', or is "localhost");
// a missing tag defaults to "latest".
func ParseImageRef(s string) (ImageRef, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ImageRef{}, errordef.Errorf(errordef.InvalidInput, "image ref is empty")
	}

	registry, remainder := splitRegistry(s)

	repository := remainder
	tag := defaultTag
	if i := strings.LastIndexByte(remainder, ':'); i >= 0 && !strings.Contains(remainder[i+1:], "/") {
		repository, tag = remainder[:i], remainder[i+1:]
	}

	if repository == "" {
		return ImageRef{}, errordef.Errorf(errordef.InvalidInput, "image ref %q has empty repository", s)
	}
	if tag == "" {
		return ImageRef{}, errordef.Errorf(errordef.InvalidInput, "image ref %q has empty tag", s)
	}
	return ImageRef{registry: registry, repository: repository, tag: tag}, nil
}

func splitRegistry(s string) (registry, remainder string) {
	i := strings.IndexByte(s, '/')
	if i < 0 {
		return "", s
	}
	if first := s[:i]; strings.ContainsAny(first, ".:") || first == "localhost" {
		return first, s[i+1:]
	}
	return "", s
}

// Registry returns the registry host, or "" when none was specified.
func (r ImageRef) Registry() string { return r.registry }

// Repository returns the repository path.
func (r ImageRef) Repository() string { return r.repository }

// Tag returns the tag.
func (r ImageRef) Tag() string { return r.tag }

// String returns the canonical [registry/]repository:tag form.
func (r ImageRef) String() string {
	if r == (ImageRef{}) {
		return ""
	}
	var b strings.Builder
	if r.registry != "" {
		b.WriteString(r.registry)
		b.WriteByte('/')
	}
	b.WriteString(r.repository)
	b.WriteByte(':')
	b.WriteString(r.tag)
	return b.String()
}

// MarshalJSON encodes the ImageRef as its canonical string.
func (r ImageRef) MarshalJSON() ([]byte, error) {
	s, err := encoding.MarshalJSON(r.String())
	return []byte(s), err
}

// UnmarshalJSON decodes an ImageRef from its string form; "" yields the zero value.
func (r *ImageRef) UnmarshalJSON(data []byte) error {
	var s string
	if err := encoding.UnmarshalJSON(string(data), &s); err != nil {
		return err
	}
	if s == "" {
		*r = ImageRef{}
		return nil
	}
	parsed, err := ParseImageRef(s)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}
