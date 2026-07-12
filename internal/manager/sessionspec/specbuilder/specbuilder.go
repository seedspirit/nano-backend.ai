package specbuilder

import (
	"context"

	"github.com/google/uuid"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/draft"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/preset"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/spec"
	"github.com/seedspirit/nano-backend.ai/internal/common/errordef"
)

// Builder turns a submitted RunDraft into an immutable Spec.
type Builder interface {
	Build(ctx context.Context, sessionDraft *draft.Draft) (spec.Spec, error)
}

// PresetRegistry is the preset lookup dependency consumed by PresetBacked.
type PresetRegistry interface {
	GetMany(ctx context.Context, ids []preset.ID) (map[preset.ID]preset.Preset, error)
}

// Validator validates a draft plus resolved preset contracts
type Validator interface {
	Validate(candidate Candidate) error
}

// PresetBacked orchestrates preset lookup, validation, and finalization.
type PresetBacked struct {
	Registry  PresetRegistry
	Validator Validator
}

// Build validates and finalizes a submitted RunDraft.
func (b PresetBacked) Build(ctx context.Context, sessionDraft *draft.Draft) (spec.Spec, error) {
	if b.Registry == nil {
		return spec.Spec{}, errordef.Errorf(errordef.InvalidInput, "preset registry is nil")
	}
	if b.Validator == nil {
		return spec.Spec{}, errordef.Errorf(errordef.InvalidInput, "session spec validator is nil")
	}

	presets, err := b.readPresets(ctx, sessionDraft)
	if err != nil {
		return spec.Spec{}, err
	}
	candidate := Candidate{Draft: sessionDraft, Presets: presets}
	if err := b.Validator.Validate(candidate); err != nil {
		return spec.Spec{}, err
	}
	return FinalizeSessionSpec(candidate), nil
}

func (b PresetBacked) readPresets(ctx context.Context, sessionDraft *draft.Draft) (preset.Presets, error) {
	if sessionDraft == nil {
		return preset.Presets{}, errordef.Errorf(errordef.InvalidInput, "draft is nil")
	}
	resolved, err := b.readPreset(ctx, sessionDraft.PresetRefs)
	if err != nil {
		return preset.Presets{}, err
	}
	return preset.Presets{
		Trainer:  presetByRef(resolved, sessionDraft.PresetRefs.Trainer),
		Resource: presetByRef(resolved, sessionDraft.PresetRefs.Resource),
		Output:   presetByRef(resolved, sessionDraft.PresetRefs.Output),
	}, nil
}

// TODO: Read preset data from database and cache it in memory for efficient lookup
func (b PresetBacked) readPreset(ctx context.Context, refs preset.Refs) (map[preset.ID]preset.Preset, error) {
	ids := collectPresetIDs(refs)
	if len(ids) == 0 {
		return map[preset.ID]preset.Preset{}, nil
	}
	return b.Registry.GetMany(ctx, ids)
}

func collectPresetIDs(refs preset.Refs) []preset.ID {
	candidates := []*preset.ID{refs.Trainer, refs.Resource, refs.Output}
	ids := make([]preset.ID, 0, len(candidates))
	seen := make(map[preset.ID]struct{}, len(candidates))
	for _, ref := range candidates {
		if ref == nil || *ref == uuid.Nil {
			continue
		}
		if _, ok := seen[*ref]; ok {
			continue
		}
		seen[*ref] = struct{}{}
		ids = append(ids, *ref)
	}
	return ids
}

func presetByRef(resolved map[preset.ID]preset.Preset, ref *preset.ID) preset.Preset {
	if ref == nil || *ref == uuid.Nil {
		return nil
	}
	return resolved[*ref]
}
