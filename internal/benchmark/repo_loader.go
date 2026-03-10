package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
)

// RepoMeta holds discovered metadata about a repository.
type RepoMeta struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	AbsPath     string   `json:"absPath"`
	IsGitRepo   bool     `json:"isGitRepo"`
	Languages   []string `json:"languages,omitempty"`
	Description string   `json:"description,omitempty"`
}

// ResolveRepo resolves a Repo config to a RepoMeta with absolute path and metadata.
func ResolveRepo(cfg Repo, projectRoot string) (*RepoMeta, error) {
	absPath := cfg.Path
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(projectRoot, absPath)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("repo %s path %s: %w", cfg.Name, absPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repo %s path %s is not a directory", cfg.Name, absPath)
	}

	meta := &RepoMeta{
		Name:        cfg.Name,
		Type:        cfg.Type,
		AbsPath:     absPath,
		Description: cfg.Description,
	}

	// Check if it's a git repo.
	if _, err := os.Stat(filepath.Join(absPath, ".git")); err == nil {
		meta.IsGitRepo = true
	}

	// Detect languages by file extension sampling.
	meta.Languages = detectLanguages(absPath)

	return meta, nil
}

// detectLanguages does a shallow scan to identify primary languages.
func detectLanguages(dir string) []string {
	extCounts := map[string]int{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	// Scan top-level and one level deep.
	for _, e := range entries {
		if e.IsDir() {
			subEntries, err := os.ReadDir(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			for _, se := range subEntries {
				if !se.IsDir() {
					ext := filepath.Ext(se.Name())
					extCounts[ext]++
				}
			}
		} else {
			ext := filepath.Ext(e.Name())
			extCounts[ext]++
		}
	}

	langMap := map[string]string{
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "javascript",
		".tsx":  "typescript",
		".py":   "python",
		".go":   "go",
		".java": "java",
		".rb":   "ruby",
		".rs":   "rust",
	}

	seen := map[string]bool{}
	var langs []string
	for ext, count := range extCounts {
		if lang, ok := langMap[ext]; ok && count >= 2 && !seen[lang] {
			seen[lang] = true
			langs = append(langs, lang)
		}
	}
	return langs
}

// DiscoverRepos finds repos in a directory for auto-discovery mode.
func DiscoverRepos(dir string) ([]Repo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("discovering repos in %s: %w", dir, err)
	}

	var repos []Repo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		repos = append(repos, Repo{
			Name: e.Name(),
			Path: filepath.Join(dir, e.Name()),
			Type: "real",
		})
	}
	return repos, nil
}
