package env

import "strings"

// Merge combines host, pipeline, and step environment overrides in order of
// increasing precedence.
func Merge(host, pipelineOverride, stepOverride map[string]string) map[string]string {
	merged := make(map[string]string, len(host)+len(pipelineOverride)+len(stepOverride))

	mergeLayer := func(layer map[string]string) {
		for key, value := range layer {
			merged[strings.ToUpper(key)] = value
		}
	}

	mergeLayer(host)
	mergeLayer(pipelineOverride)
	mergeLayer(stepOverride)

	return merged
}
