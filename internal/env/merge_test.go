package env

import (
	"reflect"
	"testing"
)

func TestEnvMerge_StepOverridesPipelineOverridesHost(t *testing.T) {
	tests := []struct {
		name             string
		host             map[string]string
		pipelineOverride map[string]string
		stepOverride     map[string]string
		want             map[string]string
	}{
		{
			name:             "step overrides pipeline and pipeline overrides host",
			host:             map[string]string{"SHARED": "host", "HOST_ONLY": "host"},
			pipelineOverride: map[string]string{"SHARED": "pipeline", "PIPELINE_ONLY": "pipeline"},
			stepOverride:     map[string]string{"SHARED": "step", "STEP_ONLY": "step"},
			want:             map[string]string{"SHARED": "step", "HOST_ONLY": "host", "PIPELINE_ONLY": "pipeline", "STEP_ONLY": "step"},
		},
		{
			name:             "empty overrides preserve host",
			host:             map[string]string{"HOST_ONLY": "host"},
			pipelineOverride: map[string]string{},
			stepOverride:     map[string]string{},
			want:             map[string]string{"HOST_ONLY": "host"},
		},
		{
			name:             "case-insensitive override replaces host",
			host:             map[string]string{"Path": "host"},
			pipelineOverride: map[string]string{"PATH": "pipeline"},
			stepOverride:     map[string]string{},
			want:             map[string]string{"PATH": "pipeline"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Merge(test.host, test.pipelineOverride, test.stepOverride); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Merge() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestEnvMerge_CaseVariantKeysAcrossLayersRespectPrecedence(t *testing.T) {
	got := Merge(
		map[string]string{"path": "host"},
		map[string]string{"PATH": "pipeline"},
		map[string]string{"Path": "step"},
	)
	want := map[string]string{"PATH": "step"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge() = %#v, want %#v", got, want)
	}
}

func TestEnvMerge_CaseVariantKeysInSameLayerHaveStableWinner(t *testing.T) {
	layer := map[string]string{
		"PATH": "upper",
		"Path": "title",
	}

	for iteration := 0; iteration < 1000; iteration++ {
		got := Merge(layer, nil, nil)
		if got["PATH"] != "title" {
			t.Fatalf("iteration %d: Merge() = %#v, want deterministic lexical winner title", iteration, got)
		}
	}
}
