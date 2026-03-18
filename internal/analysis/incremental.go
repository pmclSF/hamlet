package analysis

import (
	"sort"
	"time"

	"github.com/pmclSF/terrain/internal/models"
)

// IncrementalState holds state from a previous analysis run that can be
// reused to accelerate subsequent runs. The key optimization: files that
// haven't changed since the last run don't need re-parsing.
type IncrementalState struct {
	// Cache is the file content and AST cache from the previous run.
	// Entries are invalidated when the underlying file's mtime changes.
	Cache *FileCache

	// PreviousSnapshot is the snapshot from the last run. Used to carry
	// forward surfaces, code units, and fixtures for unchanged files.
	PreviousSnapshot *models.TestSuiteSnapshot

	// Timestamp records when the previous analysis completed.
	Timestamp time.Time

	// ChangedFiles is the set of files that changed since the last run.
	// Populated by InvalidateAndDetectChanges().
	ChangedFiles []string
}

// InvalidateAndDetectChanges checks which cached files have been modified
// on disk and invalidates their cache entries. Returns the list of changed
// file paths so the caller can decide what to re-analyze.
func (is *IncrementalState) InvalidateAndDetectChanges() []string {
	if is.Cache == nil {
		return nil
	}
	is.ChangedFiles = is.Cache.InvalidateStale()
	sort.Strings(is.ChangedFiles)
	return is.ChangedFiles
}

// ShouldReanalyze returns true if a file needs re-analysis. Files not in
// the changed set can reuse results from the previous snapshot.
func (is *IncrementalState) ShouldReanalyze(relPath string) bool {
	if is == nil || len(is.ChangedFiles) == 0 {
		return true // No incremental state — analyze everything.
	}
	idx := sort.SearchStrings(is.ChangedFiles, relPath)
	return idx < len(is.ChangedFiles) && is.ChangedFiles[idx] == relPath
}

// CarryForwardSurfaces returns code surfaces from the previous snapshot
// that belong to unchanged files. These don't need re-detection.
func (is *IncrementalState) CarryForwardSurfaces() []models.CodeSurface {
	if is == nil || is.PreviousSnapshot == nil {
		return nil
	}
	var carried []models.CodeSurface
	for _, cs := range is.PreviousSnapshot.CodeSurfaces {
		if !is.ShouldReanalyze(cs.Path) {
			carried = append(carried, cs)
		}
	}
	return carried
}

// CarryForwardCodeUnits returns code units from unchanged files.
func (is *IncrementalState) CarryForwardCodeUnits() []models.CodeUnit {
	if is == nil || is.PreviousSnapshot == nil {
		return nil
	}
	var carried []models.CodeUnit
	for _, cu := range is.PreviousSnapshot.CodeUnits {
		if !is.ShouldReanalyze(cu.Path) {
			carried = append(carried, cu)
		}
	}
	return carried
}

// CarryForwardFixtures returns fixture surfaces from unchanged test files.
func (is *IncrementalState) CarryForwardFixtures() []models.FixtureSurface {
	if is == nil || is.PreviousSnapshot == nil {
		return nil
	}
	var carried []models.FixtureSurface
	for _, fs := range is.PreviousSnapshot.FixtureSurfaces {
		if !is.ShouldReanalyze(fs.Path) {
			carried = append(carried, fs)
		}
	}
	return carried
}

// ChangedSourceFiles returns only the source files (from the provided list)
// that need re-analysis. Unchanged files are filtered out.
func (is *IncrementalState) ChangedSourceFiles(allFiles []string) []string {
	if is == nil || len(is.ChangedFiles) == 0 {
		return allFiles // No incremental state — analyze everything.
	}
	var changed []string
	for _, f := range allFiles {
		if is.ShouldReanalyze(f) {
			changed = append(changed, f)
		}
	}
	return changed
}
