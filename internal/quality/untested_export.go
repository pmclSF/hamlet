package quality

import (
	"path/filepath"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// UntestedExportDetector identifies exported/public code units that have
// no linked test coverage in the current analysis model.
//
// Heuristic:
//   - An exported code unit is considered "untested" if no test file
//     references the same directory or a matching file name pattern.
//   - This is a heuristic linkage model, not coverage-based.
//     It approximates whether any test file is "nearby" the code unit.
//
// Limitations:
//   - Cannot determine actual runtime coverage without coverage data.
//   - May produce false positives for code tested via integration tests
//     in a different directory.
//   - Heuristic linkage may miss tests that import from barrel/index files.
//
// Confidence is set lower (0.5) to reflect the heuristic nature.
type UntestedExportDetector struct{}

// Detect scans code units for untested exports.
func (d *UntestedExportDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal

	if len(snap.CodeUnits) == 0 {
		return nil
	}

	// Build a set of directories and base-name stems covered by test files.
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

		// Check if any test covers this code unit's directory or name
		cuDir := filepath.Dir(cu.Path)
		cuStem := stripExt(filepath.Base(cu.Path))

		hasNearbyTest := testDirs[cuDir] || testStems[cuStem]

		if !hasNearbyTest {
			signals = append(signals, models.Signal{
				Type:             "untestedExport",
				Category:         models.CategoryQuality,
				Severity:         models.SeverityMedium,
				Confidence:       0.5,
				EvidenceStrength: models.EvidenceWeak,
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
	if strings.HasPrefix(name, "test_") {
		name = strings.TrimPrefix(name, "test_")
	}
	return name
}

func stripExt(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}
