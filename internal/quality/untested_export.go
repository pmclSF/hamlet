package quality

import (
	"path/filepath"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// UntestedExportDetector identifies exported/public code units that have
// no linked test coverage in the current analysis model.
//
// Detection uses a layered approach:
//  1. Import graph (highest confidence): if the snapshot includes an import graph,
//     check whether any test file imports the code unit's module. This is precise
//     because it traces actual import/require statements.
//  2. Heuristic fallback (lower confidence): if no import graph is available or
//     the module isn't found in it, fall back to directory/filename-stem proximity.
//
// Limitations:
//   - Import graph only traces static, relative imports. Dynamic imports, path
//     aliases, and barrel re-exports may not be fully resolved.
//   - Heuristic fallback cannot determine actual runtime coverage.
//   - Code tested via integration tests in a different directory may be flagged
//     unless the import graph captures the linkage.
type UntestedExportDetector struct{}

// Detect scans code units for untested exports.
func (d *UntestedExportDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal

	if len(snap.CodeUnits) == 0 {
		return nil
	}

	// Layer 1: Build set of source modules imported by test files.
	importedModules := map[string]bool{}
	if snap.ImportGraph != nil {
		for _, imports := range snap.ImportGraph {
			for mod := range imports {
				importedModules[mod] = true
			}
		}
	}
	hasImportGraph := len(importedModules) > 0

	// Layer 2: Build heuristic sets (directories and filename stems).
	testDirs := map[string]bool{}
	testStems := map[string]bool{}
	for _, tf := range snap.TestFiles {
		dir := filepath.Dir(tf.Path)
		testDirs[dir] = true
		// Also consider parent dir (for __tests__/ convention)
		if filepath.Base(dir) == "__tests__" {
			testDirs[filepath.Dir(dir)] = true
		}

		// Extract stem: auth.test.js -> auth
		base := filepath.Base(tf.Path)
		stem := stripTestSuffix(base)
		if stem != "" {
			testStems[stem] = true
		}
	}

	for _, cu := range snap.CodeUnits {
		if !cu.Exported {
			continue
		}

		cuPath := filepath.ToSlash(cu.Path)
		cuDir := filepath.Dir(cuPath)
		cuStem := stripExt(filepath.Base(cuPath))

		// Layer 1: Check import graph — if any test imports this module, it's tested.
		if hasImportGraph && importedModules[cuPath] {
			continue // Tested via direct import — high confidence, no signal.
		}

		// Layer 2: Heuristic — check directory/stem proximity.
		hasNearbyTest := testDirs[cuDir] || testStems[cuStem]

		if hasNearbyTest {
			continue // Heuristic says it's likely tested.
		}

		// Determine confidence based on what evidence we have.
		confidence := 0.5
		evidenceStrength := models.EvidenceWeak
		if hasImportGraph {
			// Import graph was available but didn't find a link — higher confidence
			// that this is genuinely untested.
			confidence = 0.7
			evidenceStrength = models.EvidenceModerate
		}

		signals = append(signals, models.Signal{
			Type:             "untestedExport",
			Category:         models.CategoryQuality,
			Severity:         models.SeverityMedium,
			Confidence:       confidence,
			EvidenceStrength: evidenceStrength,
			EvidenceSource:   models.SourcePathName,
			Location: models.SignalLocation{
				File:   cu.Path,
				Symbol: cu.Name,
			},
			Explanation: "Exported " + string(cu.Kind) + " \"" + cu.Name +
				"\" has no linked tests in the current analysis model.",
			SuggestedAction: "Add direct tests for this exported behavior or improve test-to-code linkage.",
		})
	}

	return signals
}

// stripTestSuffix removes test/spec suffixes to get the base module name.
// "auth.test.js" -> "auth", "auth.spec.ts" -> "auth"
func stripTestSuffix(filename string) string {
	name := stripExt(filename)
	name = strings.TrimSuffix(name, ".test")
	name = strings.TrimSuffix(name, ".spec")
	name = strings.TrimSuffix(name, "_test")
	name = strings.TrimPrefix(name, "test_")
	return name
}

func stripExt(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}
