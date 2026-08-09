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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Merge(test.host, test.pipelineOverride, test.stepOverride); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Merge() = %#v, want %#v", got, test.want)
			}
		})
	}
}
