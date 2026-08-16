package env

import (
	"sort"
	"strings"
)

// Merge combines host, pipeline, and step environment overrides in order of
// increasing precedence.
func Merge(host, pipelineOverride, stepOverride map[string]string) map[string]string {
	merged := make(map[string]string, len(host)+len(pipelineOverride)+len(stepOverride))

	mergeLayer := func(layer map[string]string) {
		// A single layer may contain case variants of the same Windows
		// environment key. Sort before normalizing so their tie-break is stable.
		keys := make([]string, 0, len(layer))
		for key := range layer {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			merged[strings.ToUpper(key)] = layer[key]
		}
	}

	mergeLayer(host)
	mergeLayer(pipelineOverride)
	mergeLayer(stepOverride)

	return merged
}
