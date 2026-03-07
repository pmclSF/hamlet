package policy

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PolicyFileName is the expected policy file path relative to repo root.
const PolicyFileName = ".hamlet/policy.yaml"

// LoadResult contains the outcome of attempting to load a policy file.
type LoadResult struct {
	// Config is the parsed policy config. Nil if no file was found.
	Config *Config

	// Path is the absolute path to the policy file, if found.
	Path string

	// Found indicates whether a policy file existed.
	Found bool
}

// Load reads .hamlet/policy.yaml from the given repository root.
//
// Behavior:
//   - If the file does not exist, returns a LoadResult with Found=false and no error.
//   - If the file exists but is malformed, returns an actionable error.
//   - If the file is valid, returns the parsed Config.
func Load(repoRoot string) (*LoadResult, error) {
	path := filepath.Join(repoRoot, PolicyFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &LoadResult{Found: false}, nil
		}
		return nil, fmt.Errorf("failed to read policy file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("malformed policy file %s: %w", path, err)
	}

	return &LoadResult{
		Config: &cfg,
		Path:   path,
		Found:  true,
	}, nil
}
