package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
	"gopkg.in/yaml.v3"
)

// CIMatrixResult holds environments, classes, and provenance from CI config parsing.
type CIMatrixResult struct {
	Environments       []models.Environment
	EnvironmentClasses []models.EnvironmentClass
}

// ParseCIMatrices scans the repository root for CI configuration files and
// extracts environment matrices. Returns environments and classes ready to
// be added to the snapshot.
func ParseCIMatrices(root string) *CIMatrixResult {
	result := &CIMatrixResult{}

	parseGitHubActions(root, result)
	parseGitLabCI(root, result)
	parseCircleCI(root, result)
	parseBuildkite(root, result)

	return result
}

// --- GitHub Actions ---

// githubWorkflow is a minimal representation of a GitHub Actions workflow.
type githubWorkflow struct {
	Jobs map[string]githubJob `yaml:"jobs"`
}

type githubJob struct {
	RunsOn   interface{}            `yaml:"runs-on"`
	Strategy *githubStrategy        `yaml:"strategy"`
	Steps    []map[string]interface{} `yaml:"steps"`
}

type githubStrategy struct {
	Matrix    map[string]interface{} `yaml:"matrix"`
	FailFast  *bool                  `yaml:"fail-fast"`
	MaxParallel *int                 `yaml:"max-parallel"`
}

func parseGitHubActions(root string, result *CIMatrixResult) {
	workflowDir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(workflowDir, entry.Name()))
		if err != nil {
			continue
		}
		parseGitHubWorkflow(data, entry.Name(), result)
	}
}

func parseGitHubWorkflow(data []byte, filename string, result *CIMatrixResult) {
	var wf githubWorkflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return
	}

	provenance := "github-actions:" + filename

	for _, job := range wf.Jobs {
		// Extract runs-on as environment.
		if runsOn := extractRunsOn(job.RunsOn); runsOn != "" {
			envID := "env:gha-" + sanitizeID(runsOn)
			result.Environments = appendEnvIfNew(result.Environments, models.Environment{
				EnvironmentID: envID,
				Name:          runsOn,
				OS:            inferOSFromRunner(runsOn),
				CIProvider:    "github-actions",
				InferredFrom:  provenance,
			})
		}

		// Extract strategy.matrix dimensions.
		if job.Strategy == nil || job.Strategy.Matrix == nil {
			continue
		}

		for key, val := range job.Strategy.Matrix {
			// Skip special keys.
			if key == "include" || key == "exclude" {
				continue
			}
			values := extractStringSlice(val)
			if len(values) == 0 {
				continue
			}

			dimension := inferDimension(key)
			classID := "envclass:gha-" + sanitizeID(key)
			memberIDs := make([]string, 0, len(values))

			for _, v := range values {
				envID := "env:gha-" + sanitizeID(key) + "-" + sanitizeID(v)
				memberIDs = append(memberIDs, envID)

				env := models.Environment{
					EnvironmentID: envID,
					Name:          fmt.Sprintf("%s %s", key, v),
					CIProvider:    "github-actions",
					ClassID:       classID,
					InferredFrom:  provenance,
				}

				// Populate fields based on dimension.
				switch dimension {
				case "os":
					env.OS = inferOSFromRunner(v)
					env.Name = v
				case "runtime":
					env.Runtime = v
					env.Name = key + " " + v
				}

				result.Environments = appendEnvIfNew(result.Environments, env)
			}

			result.EnvironmentClasses = appendClassIfNew(result.EnvironmentClasses, models.EnvironmentClass{
				ClassID:   classID,
				Name:      key,
				Dimension: dimension,
				MemberIDs: memberIDs,
			})
		}
	}
}

func extractRunsOn(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		if len(val) > 0 {
			if s, ok := val[0].(string); ok {
				return s
			}
		}
	}
	return ""
}

// --- GitLab CI ---

func parseGitLabCI(root string, result *CIMatrixResult) {
	data, err := os.ReadFile(filepath.Join(root, ".gitlab-ci.yml"))
	if err != nil {
		return
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return
	}

	provenance := "gitlab-ci:.gitlab-ci.yml"

	for jobName, jobVal := range raw {
		// Skip GitLab CI top-level keywords.
		if isGitLabKeyword(jobName) {
			continue
		}
		jobMap, ok := jobVal.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract parallel:matrix.
		parallel, ok := jobMap["parallel"]
		if !ok {
			continue
		}
		parallelMap, ok := parallel.(map[string]interface{})
		if !ok {
			continue
		}
		matrixVal, ok := parallelMap["matrix"]
		if !ok {
			continue
		}

		// matrix is a list of variable-set maps.
		matrixList, ok := matrixVal.([]interface{})
		if !ok {
			continue
		}

		for _, entry := range matrixList {
			entryMap, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			for key, val := range entryMap {
				values := extractStringSlice(val)
				if len(values) == 0 {
					continue
				}

				dimension := inferDimension(key)
				classID := "envclass:gitlab-" + sanitizeID(key)
				memberIDs := make([]string, 0, len(values))

				for _, v := range values {
					envID := "env:gitlab-" + sanitizeID(key) + "-" + sanitizeID(v)
					memberIDs = append(memberIDs, envID)

					result.Environments = appendEnvIfNew(result.Environments, models.Environment{
						EnvironmentID: envID,
						Name:          key + " " + v,
						CIProvider:    "gitlab-ci",
						ClassID:       classID,
						InferredFrom:  provenance,
					})
				}

				result.EnvironmentClasses = appendClassIfNew(result.EnvironmentClasses, models.EnvironmentClass{
					ClassID:   classID,
					Name:      key,
					Dimension: dimension,
					MemberIDs: memberIDs,
				})
			}
		}
	}
}

func isGitLabKeyword(name string) bool {
	keywords := map[string]bool{
		"stages": true, "variables": true, "default": true,
		"include": true, "workflow": true, "image": true,
		"services": true, "before_script": true, "after_script": true,
		"cache": true, "artifacts": true,
	}
	return keywords[name] || strings.HasPrefix(name, ".")
}

// --- CircleCI ---

func parseCircleCI(root string, result *CIMatrixResult) {
	data, err := os.ReadFile(filepath.Join(root, ".circleci", "config.yml"))
	if err != nil {
		return
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return
	}

	provenance := "circleci:.circleci/config.yml"

	// Parse workflows for matrix jobs.
	workflows, ok := raw["workflows"].(map[string]interface{})
	if !ok {
		return
	}

	for _, wfVal := range workflows {
		wfMap, ok := wfVal.(map[string]interface{})
		if !ok {
			continue
		}
		jobs, ok := wfMap["jobs"].([]interface{})
		if !ok {
			continue
		}
		for _, jobEntry := range jobs {
			jobMap, ok := jobEntry.(map[string]interface{})
			if !ok {
				continue
			}
			for _, jobConfig := range jobMap {
				cfgMap, ok := jobConfig.(map[string]interface{})
				if !ok {
					continue
				}
				matrixVal, ok := cfgMap["matrix"]
				if !ok {
					continue
				}
				matrixMap, ok := matrixVal.(map[string]interface{})
				if !ok {
					continue
				}
				params, ok := matrixMap["parameters"].(map[string]interface{})
				if !ok {
					continue
				}

				for key, val := range params {
					values := extractStringSlice(val)
					if len(values) == 0 {
						continue
					}

					dimension := inferDimension(key)
					classID := "envclass:circleci-" + sanitizeID(key)
					memberIDs := make([]string, 0, len(values))

					for _, v := range values {
						envID := "env:circleci-" + sanitizeID(key) + "-" + sanitizeID(v)
						memberIDs = append(memberIDs, envID)

						env := models.Environment{
							EnvironmentID: envID,
							Name:          key + " " + v,
							CIProvider:    "circleci",
							ClassID:       classID,
							InferredFrom:  provenance,
						}
						if dimension == "runtime" {
							env.Runtime = v
						}

						result.Environments = appendEnvIfNew(result.Environments, env)
					}

					result.EnvironmentClasses = appendClassIfNew(result.EnvironmentClasses, models.EnvironmentClass{
						ClassID:   classID,
						Name:      key,
						Dimension: dimension,
						MemberIDs: memberIDs,
					})
				}
			}
		}
	}
}

// --- Buildkite ---

func parseBuildkite(root string, result *CIMatrixResult) {
	// Try .buildkite/pipeline.yml.
	paths := []string{
		filepath.Join(root, ".buildkite", "pipeline.yml"),
		filepath.Join(root, ".buildkite", "pipeline.yaml"),
	}

	var data []byte
	var provFile string
	for _, p := range paths {
		d, err := os.ReadFile(p)
		if err == nil {
			data = d
			provFile = strings.TrimPrefix(p, root+string(filepath.Separator))
			break
		}
	}
	if data == nil {
		return
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return
	}

	provenance := "buildkite:" + provFile

	// Buildkite matrix is under steps[].matrix.setup.
	steps, ok := raw["steps"].([]interface{})
	if !ok {
		return
	}

	for _, step := range steps {
		stepMap, ok := step.(map[string]interface{})
		if !ok {
			continue
		}
		matrixVal, ok := stepMap["matrix"]
		if !ok {
			continue
		}

		// matrix can be a simple list or a map with "setup" key.
		var setupMap map[string]interface{}
		switch m := matrixVal.(type) {
		case map[string]interface{}:
			if setup, ok := m["setup"].(map[string]interface{}); ok {
				setupMap = setup
			} else {
				setupMap = m
			}
		default:
			continue
		}

		for key, val := range setupMap {
			values := extractStringSlice(val)
			if len(values) == 0 {
				continue
			}

			dimension := inferDimension(key)
			classID := "envclass:buildkite-" + sanitizeID(key)
			memberIDs := make([]string, 0, len(values))

			for _, v := range values {
				envID := "env:buildkite-" + sanitizeID(key) + "-" + sanitizeID(v)
				memberIDs = append(memberIDs, envID)

				result.Environments = appendEnvIfNew(result.Environments, models.Environment{
					EnvironmentID: envID,
					Name:          key + " " + v,
					CIProvider:    "buildkite",
					ClassID:       classID,
					InferredFrom:  provenance,
				})
			}

			result.EnvironmentClasses = appendClassIfNew(result.EnvironmentClasses, models.EnvironmentClass{
				ClassID:   classID,
				Name:      key,
				Dimension: dimension,
				MemberIDs: memberIDs,
			})
		}
	}
}

// --- Helpers ---

// sanitizeID normalizes a value for use in a stable ID.
func sanitizeID(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			return r
		}
		return '-'
	}, s)
	// Collapse multiple dashes.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}

// inferDimension guesses the dimension from a matrix key name.
func inferDimension(key string) string {
	lower := strings.ToLower(key)
	switch {
	case lower == "os" || lower == "runs-on" || lower == "platform" || lower == "runner":
		return "os"
	case lower == "node-version" || lower == "node_version" || lower == "node" ||
		lower == "python-version" || lower == "python_version" || lower == "python" ||
		lower == "go-version" || lower == "go_version" || lower == "go" ||
		lower == "java-version" || lower == "java_version" || lower == "java" ||
		lower == "ruby-version" || lower == "ruby_version" || lower == "ruby" ||
		lower == "version" || lower == "runtime":
		return "runtime"
	case lower == "browser" || lower == "browsers":
		return "browser"
	case lower == "device" || lower == "devices":
		return "device"
	case lower == "arch" || lower == "architecture":
		return "architecture"
	case lower == "region":
		return "region"
	default:
		return lower
	}
}

// inferOSFromRunner extracts the OS from a GitHub Actions runner label.
func inferOSFromRunner(runner string) string {
	lower := strings.ToLower(runner)
	switch {
	case strings.Contains(lower, "ubuntu") || strings.Contains(lower, "linux"):
		return "linux"
	case strings.Contains(lower, "macos") || strings.Contains(lower, "mac"):
		return "macos"
	case strings.Contains(lower, "windows"):
		return "windows"
	default:
		return ""
	}
}

// extractStringSlice normalizes a YAML value to a string slice.
func extractStringSlice(v interface{}) []string {
	switch val := v.(type) {
	case []interface{}:
		var out []string
		for _, item := range val {
			s := fmt.Sprintf("%v", item)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		return []string{val}
	default:
		s := fmt.Sprintf("%v", val)
		if s != "" && s != "<nil>" {
			return []string{s}
		}
		return nil
	}
}

func appendEnvIfNew(envs []models.Environment, env models.Environment) []models.Environment {
	for _, e := range envs {
		if e.EnvironmentID == env.EnvironmentID {
			return envs
		}
	}
	return append(envs, env)
}

func appendClassIfNew(classes []models.EnvironmentClass, cls models.EnvironmentClass) []models.EnvironmentClass {
	for i, c := range classes {
		if c.ClassID == cls.ClassID {
			// Merge member IDs.
			seen := map[string]bool{}
			for _, id := range c.MemberIDs {
				seen[id] = true
			}
			for _, id := range cls.MemberIDs {
				if !seen[id] {
					classes[i].MemberIDs = append(classes[i].MemberIDs, id)
				}
			}
			sort.Strings(classes[i].MemberIDs)
			return classes
		}
	}
	return append(classes, cls)
}
