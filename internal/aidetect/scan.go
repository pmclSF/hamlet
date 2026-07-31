package aidetect

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/gitignore"
	"github.com/pmclSF/terrain/internal/models"
)

// skipDirs are directories the AI-config walker never descends into.
// These MUST match the canonical set in
// internal/analysis/repository_scan.go — drift here causes detectors to
// re-scan trees other walkers correctly avoid. In particular, descending
// into .terrain/ would re-detect the engine's own previously-saved
// snapshots and inflate signal counts on each successive
// `terrain analyze --write-snapshot` run.
var skipDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"dist":          true,
	"build":         true,
	"benchmarks":    true,
	"coverage":      true,
	".next":         true,
	".turbo":        true,
	".nuxt":         true,
	"vendor":        true,
	"__pycache__":   true,
	".pytest_cache": true,
	".mypy_cache":   true,
	".tox":          true,
	".venv":         true,
	"venv":          true,
	".idea":         true,
	".vscode":       true,
	".terrain":      true,
	"target":        true,
}

// scanOpts tunes the walker. Detectors compose their narrow allowlist
// (extensions + filename markers) and pass it in.
type scanOpts struct {
	// extensions is the set of lowercase file extensions to consider
	// (e.g. ".yaml", ".json"). Empty = match everything.
	extensions map[string]bool
	// markers is a list of substring markers; at least one must appear
	// in the file's lowercase relative path for the file to be returned.
	// Empty = no filename marker filter.
	markers []string
}

// walkRepoForConfigs walks root and returns repo-relative paths whose
// extension+filename match the given options. Skips known noisy
// directories. Returns paths in deterministic (filepath.Walk) order,
// which is sorted by directory entry name on most OSes.
func walkRepoForConfigs(root string, opts scanOpts) []string {
	var out []string
	if root == "" {
		return out
	}
	// Honour the repo .gitignore, matching the analysis walker — a config
	// inside a gitignored tree is not the repo's own surface.
	ign := gitignore.Load(root)
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if rel, relErr := filepath.Rel(root, path); relErr == nil && ign.Match(filepath.ToSlash(rel), true) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if len(opts.extensions) > 0 && !opts.extensions[ext] {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if ign.Match(rel, false) {
			return nil
		}
		if len(opts.markers) > 0 {
			lower := strings.ToLower(rel)
			matched := false
			for _, m := range opts.markers {
				if strings.Contains(lower, m) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}
		out = append(out, rel)
		return nil
	})
	return out
}

// uniquePaths merges N path lists into one with stable ordering and
// duplicate suppression. Used by detectors that combine the snapshot's
// TestFiles / Scenarios with a fresh repo walk.
func uniquePaths(lists ...[]string) []string {
	seen := map[string]bool{}
	var out []string
	for _, list := range lists {
		for _, p := range list {
			if seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// snapshotPaths pulls TestFile and Scenario paths from a snapshot.
// Helper used alongside walkRepoForConfigs by every AI detector.
func snapshotPaths(snap *models.TestSuiteSnapshot) []string {
	if snap == nil {
		return nil
	}
	out := make([]string, 0, len(snap.TestFiles)+len(snap.Evals))
	for _, tf := range snap.TestFiles {
		out = append(out, tf.Path)
	}
	for _, sc := range snap.Evals {
		if sc.Path != "" {
			out = append(out, sc.Path)
		}
	}
	return out
}

// pairedConfidence scales confidence by the number of paired cases
// behind a regression inference. A drift over 1 paired case is much
// weaker evidence than the same drift over 100 — consumers shouldn't
// see both at the same alarm level. Used by aiCostRegression and
// aiRetrievalRegression.
//
// Curve: 0.5 at paired=1, 0.7 at paired=5, 0.85 at paired=10, plateau
// at 0.9 from paired>=20. Linear interpolation inside each band keeps
// the function easy to reason about and matches the rough "you need
// double-digit case counts before a regression call is high-confidence"
// intuition.
func pairedConfidence(paired int) float64 {
	switch {
	case paired <= 1:
		return 0.5
	case paired < 5:
		return 0.5 + 0.2*float64(paired-1)/4
	case paired < 10:
		return 0.7 + 0.15*float64(paired-5)/5
	case paired < 20:
		return 0.85 + 0.05*float64(paired-10)/10
	default:
		return 0.9
	}
}
