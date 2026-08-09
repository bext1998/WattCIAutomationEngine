package pipeline

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultShell = "pwsh"

type PipelineFile struct {
	Version   int                 `yaml:"version"`
	Env       map[string]string   `yaml:"env"`
	Pipelines map[string]Pipeline `yaml:"pipelines"`
}

type Pipeline struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name  string            `yaml:"name"`
	Exec  string            `yaml:"exec"`
	Args  []string          `yaml:"args"`
	Run   string            `yaml:"run"`
	Shell string            `yaml:"shell"`
	Cwd   string            `yaml:"cwd"`
	Env   map[string]string `yaml:"env"`
}

// Load decodes one pipeline YAML document with strict field checking.
func Load(path string) (PipelineFile, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return PipelineFile{}, fmt.Errorf("load pipeline %q: %w", path, err)
	}

	structureDecoder := yaml.NewDecoder(bytes.NewReader(contents))
	var document yaml.Node
	if err := structureDecoder.Decode(&document); err != nil {
		if errors.Is(err, io.EOF) {
			return PipelineFile{}, fmt.Errorf("load pipeline %q: empty YAML document", path)
		}
		return PipelineFile{}, fmt.Errorf("load pipeline %q: %w", path, err)
	}
	if isNullDocument(&document) {
		return PipelineFile{}, fmt.Errorf("load pipeline %q: empty YAML document", path)
	}

	var extraDocument yaml.Node
	if err := structureDecoder.Decode(&extraDocument); err != io.EOF {
		if err == nil {
			return PipelineFile{}, fmt.Errorf("load pipeline %q: multiple YAML documents are not supported", path)
		}
		return PipelineFile{}, fmt.Errorf("load pipeline %q: %w", path, err)
	}
	if err := validateVersionType(&document); err != nil {
		return PipelineFile{}, fmt.Errorf("load pipeline %q: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)

	var file PipelineFile
	if err := decoder.Decode(&file); err != nil {
		return PipelineFile{}, fmt.Errorf("load pipeline %q: %w", path, err)
	}

	applyDefaults(&file)
	return file, nil
}

func isNullDocument(document *yaml.Node) bool {
	root := document
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) == 0 {
			return true
		}
		root = root.Content[0]
	}
	return root.Kind == yaml.ScalarNode && root.Tag == "!!null"
}

func validateVersionType(document *yaml.Node) error {
	root := document
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return nil
	}

	for index := 0; index+1 < len(root.Content); index += 2 {
		key := root.Content[index]
		if key.Value != "version" {
			continue
		}
		version := root.Content[index+1]
		if version.Kind == yaml.AliasNode && version.Alias != nil {
			version = version.Alias
		}
		if version.Tag != "!!int" {
			return fmt.Errorf("version at line %d must be an integer", version.Line)
		}
		return nil
	}

	return nil
}

// Validate checks the complete static pipeline contract.
func (file PipelineFile) Validate() error {
	if file.Version != 1 {
		return fmt.Errorf("invalid pipeline version %d: must be integer 1", file.Version)
	}
	if len(file.Pipelines) == 0 {
		return errors.New("invalid pipeline: pipelines must contain at least one named pipeline")
	}

	names := make([]string, 0, len(file.Pipelines))
	for name := range file.Pipelines {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, pipelineName := range names {
		pipeline := file.Pipelines[pipelineName]
		if strings.TrimSpace(pipelineName) == "" {
			return errors.New("invalid pipeline: pipeline name must not be blank")
		}
		if len(pipeline.Steps) == 0 {
			return fmt.Errorf("invalid pipeline %q: steps must contain at least one step", pipelineName)
		}

		seenStepNames := make(map[string]struct{}, len(pipeline.Steps))
		for index, step := range pipeline.Steps {
			stepNumber := index + 1
			stepName := strings.TrimSpace(step.Name)
			if stepName == "" {
				return fmt.Errorf("invalid pipeline %q step %d: name must not be blank", pipelineName, stepNumber)
			}
			if _, exists := seenStepNames[stepName]; exists {
				return fmt.Errorf("invalid pipeline %q step %d: duplicate step name %q", pipelineName, stepNumber, stepName)
			}
			seenStepNames[stepName] = struct{}{}

			hasExec := strings.TrimSpace(step.Exec) != ""
			hasRun := strings.TrimSpace(step.Run) != ""
			if hasExec == hasRun {
				return fmt.Errorf("invalid pipeline %q step %d %q: exec and run must specify exactly one", pipelineName, stepNumber, step.Name)
			}

			if !hasExec && len(step.Args) > 0 {
				return fmt.Errorf("invalid pipeline %q step %d %q: args is only valid with exec", pipelineName, stepNumber, step.Name)
			}
			if hasExec && step.Shell != "" {
				return fmt.Errorf("invalid pipeline %q step %d %q: shell is only valid with run", pipelineName, stepNumber, step.Name)
			}

			if hasRun {
				shell := step.Shell
				if shell == "" {
					shell = defaultShell
				}
				switch shell {
				case defaultShell, "cmd":
				case "bash":
					return fmt.Errorf("invalid pipeline %q step %d %q: shell bash 尚未支援", pipelineName, stepNumber, step.Name)
				default:
					return fmt.Errorf("invalid pipeline %q step %d %q: unsupported shell %q; MVP supports %s and cmd", pipelineName, stepNumber, step.Name, shell, defaultShell)
				}
			}
		}
	}

	return nil
}

func applyDefaults(file *PipelineFile) {
	for pipelineName, definition := range file.Pipelines {
		for index := range definition.Steps {
			step := &definition.Steps[index]
			if strings.TrimSpace(step.Run) != "" && step.Shell == "" {
				step.Shell = defaultShell
			}
		}
		file.Pipelines[pipelineName] = definition
	}
}
