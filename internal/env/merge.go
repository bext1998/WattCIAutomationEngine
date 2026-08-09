package env

// Merge combines host, pipeline, and step environment overrides in order of
// increasing precedence.
func Merge(host, pipelineOverride, stepOverride map[string]string) map[string]string {
	merged := make(map[string]string, len(host)+len(pipelineOverride)+len(stepOverride))

	for key, value := range host {
		merged[key] = value
	}
	for key, value := range pipelineOverride {
		merged[key] = value
	}
	for key, value := range stepOverride {
		merged[key] = value
	}

	return merged
}
