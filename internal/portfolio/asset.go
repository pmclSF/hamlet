package portfolio

import (
	"path/filepath"
	"sort"

	"github.com/pmclSF/terrain/internal/models"
)

// BuildAssets derives TestAsset entries from snapshot data.
// Each test file becomes one asset with cost, reach, and stability metadata.
func BuildAssets(snap *models.TestSuiteSnapshot) []TestAsset {
	if len(snap.TestFiles) == 0 {
		return nil
	}

	// Build indexes for fast lookup.
	unitsByLinked := buildLinkedUnitIndex(snap)
	signalsByFile := buildSignalIndex(snap)
	ownerByFile := buildOwnerIndex(snap)
	fwTypes := buildFrameworkTypeIndex(snap)

	assets := make([]TestAsset, 0, len(snap.TestFiles))
	for _, tf := range snap.TestFiles {
		a := TestAsset{
			Path:      tf.Path,
			Framework: tf.Framework,
			Owner:     resolveOwner(tf, ownerByFile),
			TestCount: tf.TestCount,
		}

		// Test type from framework.
		a.TestType = inferTestType(tf, fwTypes)

		// Cost metrics from runtime data.
		if tf.RuntimeStats != nil && tf.RuntimeStats.AvgRuntimeMs > 0 {
			a.RuntimeMs = tf.RuntimeStats.AvgRuntimeMs
			a.RetryRate = tf.RuntimeStats.RetryRate
			a.PassRate = tf.RuntimeStats.PassRate
			a.HasRuntimeData = true
		}

		// Instability from signals.
		a.InstabilitySignals = len(signalsByFile[tf.Path])

		// Cost classification.
		a.CostClass = classifyCost(a)

		// Protection breadth from linked code units.
		populateProtection(&a, tf, unitsByLinked, snap)

		// Import graph linkage for precise redundancy detection.
		if imports, ok := snap.ImportGraph[tf.Path]; ok && len(imports) > 0 {
			sources := make([]string, 0, len(imports))
			for src := range imports {
				sources = append(sources, src)
			}
			sort.Strings(sources)
			a.ImportedSources = sources
		}

		assets = append(assets, a)
	}

	// Sort by path for determinism.
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Path < assets[j].Path
	})

	return assets
}

func classifyCost(a TestAsset) CostClass {
	if !a.HasRuntimeData {
		// Keep no-runtime classifications conservative to avoid overstating cost.
		switch a.TestType {
		case "e2e":
			return CostModerate
		case "integration":
			return CostModerate
		default:
			return CostUnknown
		}
	}

	// Runtime-based classification.
	switch {
	case a.RuntimeMs >= 10000 || a.RetryRate >= 0.3:
		return CostHigh
	case a.RuntimeMs >= 3000 || a.RetryRate >= 0.1:
		return CostModerate
	default:
		return CostLow
	}
}

func populateProtection(a *TestAsset, tf models.TestFile, unitsByLinked map[string][]linkedUnit, snap *models.TestSuiteSnapshot) {
	linked := unitsByLinked[tf.Path]
	if len(linked) == 0 && len(tf.LinkedCodeUnits) == 0 {
		a.BreadthClass = BreadthUnknown
		return
	}

	a.HasCoverageData = true
	a.CoveredUnitCount = len(linked)

	// If no linked units resolved but we have names, use name count.
	if a.CoveredUnitCount == 0 && len(tf.LinkedCodeUnits) > 0 {
		a.CoveredUnitCount = len(tf.LinkedCodeUnits)
	}

	// Compute modules and owners touched.
	modules := map[string]bool{}
	owners := map[string]bool{}
	exportedCount := 0

	for _, lu := range linked {
		modules[filepath.Dir(lu.Path)] = true
		if lu.Owner != "" {
			owners[lu.Owner] = true
		}
		if lu.Exported {
			exportedCount++
		}
	}

	a.ExportedUnitsCovered = exportedCount
	for m := range modules {
		a.CoveredModules = append(a.CoveredModules, m)
	}
	sort.Strings(a.CoveredModules)
	for o := range owners {
		a.OwnersCovered = append(a.OwnersCovered, o)
	}
	sort.Strings(a.OwnersCovered)

	// Breadth classification.
	moduleCount := len(modules)
	ownerCount := len(owners)

	switch {
	case moduleCount >= 3 || ownerCount >= 2:
		a.BreadthClass = BreadthBroad
	case moduleCount >= 2 || a.CoveredUnitCount >= 5:
		a.BreadthClass = BreadthModerate
	case a.CoveredUnitCount >= 1:
		a.BreadthClass = BreadthNarrow
	default:
		a.BreadthClass = BreadthUnknown
	}
}

// linkedUnit is a resolved code unit for index building.
type linkedUnit struct {
	UnitID   string
	Path     string
	Exported bool
	Owner    string
}

func buildLinkedUnitIndex(snap *models.TestSuiteSnapshot) map[string][]linkedUnit {
	// Map code unit references (UnitID and Name) to all matching code units.
	unitsByRef := map[string][]*models.CodeUnit{}
	for i := range snap.CodeUnits {
		cu := &snap.CodeUnits[i]
		if cu.Name != "" {
			unitsByRef[cu.Name] = append(unitsByRef[cu.Name], cu)
		}
		if cu.UnitID != "" {
			unitsByRef[cu.UnitID] = append(unitsByRef[cu.UnitID], cu)
		}
	}

	index := map[string][]linkedUnit{}
	for _, tf := range snap.TestFiles {
		seen := map[string]bool{}
		for _, ref := range tf.LinkedCodeUnits {
			candidates := unitsByRef[ref]
			for _, cu := range candidates {
				key := cu.UnitID + "|" + cu.Path
				if seen[key] {
					continue
				}
				seen[key] = true
				index[tf.Path] = append(index[tf.Path], linkedUnit{
					UnitID:   cu.UnitID,
					Path:     cu.Path,
					Exported: cu.Exported,
					Owner:    cu.Owner,
				})
			}
		}
	}
	return index
}

func buildSignalIndex(snap *models.TestSuiteSnapshot) map[string][]models.Signal {
	index := map[string][]models.Signal{}
	healthTypes := map[models.SignalType]bool{
		"slowTest": true, "flakyTest": true, "skippedTest": true,
		"deadTest": true, "unstableSuite": true,
	}
	for _, s := range snap.Signals {
		if healthTypes[s.Type] && s.Location.File != "" {
			index[s.Location.File] = append(index[s.Location.File], s)
		}
	}
	return index
}

func buildOwnerIndex(snap *models.TestSuiteSnapshot) map[string]string {
	index := map[string]string{}
	for path, owners := range snap.Ownership {
		if len(owners) > 0 {
			index[path] = owners[0]
		}
	}
	return index
}

func buildFrameworkTypeIndex(snap *models.TestSuiteSnapshot) map[string]models.FrameworkType {
	index := map[string]models.FrameworkType{}
	for _, fw := range snap.Frameworks {
		index[fw.Name] = fw.Type
	}
	return index
}

func resolveOwner(tf models.TestFile, ownerByFile map[string]string) string {
	if tf.Owner != "" {
		return tf.Owner
	}
	if o, ok := ownerByFile[tf.Path]; ok {
		return o
	}
	return "unknown"
}

func inferTestType(tf models.TestFile, fwTypes map[string]models.FrameworkType) string {
	ft := fwTypes[tf.Framework]
	switch ft {
	case models.FrameworkTypeUnit:
		return "unit"
	case models.FrameworkTypeIntegration:
		return "integration"
	case models.FrameworkTypeE2E:
		return "e2e"
	default:
		return "unknown"
	}
}
