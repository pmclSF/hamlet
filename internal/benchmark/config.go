package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
)

// Repo describes a repository to benchmark.
type Repo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "real" or "fixture"
	Description string `json:"description,omitempty"`
}

// BenchConfig holds the full benchmark configuration.
type BenchConfig struct {
	Repos []Repo `json:"repos"`
}

// LoadBenchmarkRepos reads the benchmark config from a JSON file and returns
// the list of configured repositories.
func LoadBenchmarkRepos(configPath string) ([]Repo, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", configPath, err)
	}
	var cfg BenchConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", configPath, err)
	}
	return cfg.Repos, nil
}
