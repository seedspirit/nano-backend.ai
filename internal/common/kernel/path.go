package kernel

import (
	"strings"

	"github.com/seedspirit/nano-backend.ai/internal/common/encoding"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// AgentPath is an opaque location on the agent host, kept distinct from
// manager-local filesystem paths. The zero value is invalid; use NewAgentPath.
type AgentPath struct {
	value string
}

// NewAgentPath constructs an AgentPath, rejecting empty input.
func NewAgentPath(s string) (AgentPath, error) {
	if strings.TrimSpace(s) == "" {
		return AgentPath{}, errordef.Errorf(errordef.InvalidInput, "agent path is empty")
	}
	return AgentPath{value: s}, nil
}

// String returns the underlying path.
func (p AgentPath) String() string { return p.value }

// MarshalJSON encodes the AgentPath as a string.
func (p AgentPath) MarshalJSON() ([]byte, error) {
	s, err := encoding.MarshalJSON(p.value)
	return []byte(s), err
}

// UnmarshalJSON decodes an AgentPath from a string; "" yields the zero value.
func (p *AgentPath) UnmarshalJSON(data []byte) error {
	var s string
	if err := encoding.UnmarshalJSON(string(data), &s); err != nil {
		return err
	}
	if s == "" {
		*p = AgentPath{}
		return nil
	}
	parsed, err := NewAgentPath(s)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}
