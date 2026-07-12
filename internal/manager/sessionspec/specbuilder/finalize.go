package specbuilder

import (
	"encoding/json"
	"fmt"

	"github.com/seedspirit/nano-backend.ai/internal/common/data/session"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/draft"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/preset"
	"github.com/seedspirit/nano-backend.ai/internal/common/data/session/spec"
)

// FinalizeSessionSpec combines a validated candidate into an immutable session Spec.
func FinalizeSessionSpec(candidate Candidate) spec.Spec {
	return overridePresets(candidate)
}

func overridePresets(candidate Candidate) spec.Spec {
	sessionDraft := candidate.Draft
	sessionSpec := spec.Spec{
		ID:              sessionDraft.ID,
		ProjectID:       sessionDraft.ProjectID,
		Name:            sessionDraft.Name,
		Description:     sessionDraft.Description,
		Type:            sessionDraft.Type,
		PresetRefs:      sessionDraft.PresetRefs,
		ModelOptions:    modelOptionsFromReq(sessionDraft.ModelOptions),
		DataOptions:     dataOptionsFromReq(sessionDraft.DataOptions),
		ResourceOptions: resourceOptionsFromReq(sessionDraft.ResourceOptions),
		TrainingOptions: session.TrainingOptions{Parameters: map[string]any{}},
	}

	overridePresetOptions(&sessionSpec, candidate.Presets)
	overrideDraftOptions(&sessionSpec, sessionDraft)

	return sessionSpec
}

func overridePresetOptions(sessionSpec *spec.Spec, presets preset.Presets) {
	for _, resolvedPreset := range presets.All() {
		if resolvedPreset == nil {
			continue
		}
		options := resolvedPreset.Options()
		if options.Model != nil {
			sessionSpec.ModelOptions = modelOptionsFromPreset(*options.Model)
		}
		if options.Data != nil {
			sessionSpec.DataOptions = dataOptionsFromPreset(*options.Data)
		}
		if options.Resource != nil {
			sessionSpec.ResourceOptions = resourceOptionsFromPreset(*options.Resource)
		}
		for key, value := range options.TrainingParameters {
			sessionSpec.TrainingOptions.Parameters[key] = copyValue(value)
		}
	}
}

func overrideDraftOptions(sessionSpec *spec.Spec, sessionDraft *draft.Draft) {
	for key, value := range sessionDraft.TrainingOptions.Parameters {
		sessionSpec.TrainingOptions.Parameters[key] = copyValue(value)
	}
}

// CanonicalJSON returns deterministic JSON for comparison and idempotency checks.
func CanonicalJSON(sessionSpec *spec.Spec) (string, error) {
	data, err := json.Marshal(sessionSpec)
	if err != nil {
		return "", fmt.Errorf("canonicalize spec %s: %w", sessionSpec.ID, err)
	}
	return string(data), nil
}

func copyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		cloned := make(map[string]any, len(typed))
		for key, item := range typed {
			cloned[key] = copyValue(item)
		}
		return cloned
	case []any:
		cloned := make([]any, len(typed))
		for i, item := range typed {
			cloned[i] = copyValue(item)
		}
		return cloned
	default:
		return value
	}
}

func modelOptionsFromReq(req draft.ModelOptionsReq) session.ModelOptions {
	return session.ModelOptions{BaseModel: req.BaseModel}
}

func dataOptionsFromReq(req draft.DataOptionsReq) session.DataOptions {
	datasets := make([]session.DatasetRef, len(req.Datasets))
	for i, item := range req.Datasets {
		datasets[i] = session.DatasetRef{Path: item.Path, Split: item.Split}
	}
	return session.DataOptions{Datasets: datasets}
}

func resourceOptionsFromReq(req draft.ResourceOptionsReq) session.ResourceOptions {
	return session.ResourceOptions{
		CPU:     req.CPU,
		GPU:     req.GPU,
		Memory:  req.Memory,
		Timeout: req.Timeout,
	}
}

func modelOptionsFromPreset(options preset.ModelOptions) session.ModelOptions {
	return session.ModelOptions{BaseModel: options.BaseModel}
}

func dataOptionsFromPreset(options preset.DataOptions) session.DataOptions {
	datasets := make([]session.DatasetRef, len(options.Datasets))
	for i, item := range options.Datasets {
		datasets[i] = session.DatasetRef{Path: item.Path, Split: item.Split}
	}
	return session.DataOptions{Datasets: datasets}
}

func resourceOptionsFromPreset(options preset.ResourceOptions) session.ResourceOptions {
	return session.ResourceOptions{
		CPU:     options.CPU,
		GPU:     options.GPU,
		Memory:  options.Memory,
		Timeout: options.Timeout,
	}
}
