package models

import "sort"

// SortSnapshot sorts all slice fields in the snapshot into a canonical,
// deterministic order. This ensures that identical inputs always produce
// byte-identical JSON output regardless of map iteration order, goroutine
// scheduling, or filesystem traversal order.
//
// Sort keys are chosen to be stable and meaningful:
//   - TestFiles: by Path
//   - TestCases: by TestID
//   - CodeUnits: by UnitID, then Path+Name
//   - Signals: by Category, Type, File, Line, Explanation
//   - Frameworks: by Name
//   - Risk: by Type, Scope, ScopeName
//   - CoverageInsights: by Type, Path, UnitID
func SortSnapshot(snap *TestSuiteSnapshot) {
	if snap == nil {
		return
	}
	normalizeCodeUnitIDs(snap)

	sort.Slice(snap.TestFiles, func(i, j int) bool {
		return snap.TestFiles[i].Path < snap.TestFiles[j].Path
	})
	for i := range snap.TestFiles {
		tf := &snap.TestFiles[i]
		if len(tf.LinkedCodeUnits) > 1 {
			sort.Strings(tf.LinkedCodeUnits)
		}
		sortSignals(tf.Signals)
	}

	sort.Slice(snap.TestCases, func(i, j int) bool {
		if snap.TestCases[i].TestID != snap.TestCases[j].TestID {
			return snap.TestCases[i].TestID < snap.TestCases[j].TestID
		}
		return snap.TestCases[i].FilePath < snap.TestCases[j].FilePath
	})

	sort.Slice(snap.CodeUnits, func(i, j int) bool {
		a, b := snap.CodeUnits[i], snap.CodeUnits[j]
		if a.UnitID != b.UnitID {
			return a.UnitID < b.UnitID
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Name < b.Name
	})

	sort.Slice(snap.Frameworks, func(i, j int) bool {
		return snap.Frameworks[i].Name < snap.Frameworks[j].Name
	})

	sortSignals(snap.Signals)

	sort.Slice(snap.Risk, func(i, j int) bool {
		a, b := snap.Risk[i], snap.Risk[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Scope != b.Scope {
			return a.Scope < b.Scope
		}
		return a.ScopeName < b.ScopeName
	})

	sort.Slice(snap.CoverageInsights, func(i, j int) bool {
		a, b := snap.CoverageInsights[i], snap.CoverageInsights[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.UnitID < b.UnitID
	})

	normalizeSnapshotMaps(snap)
}

func normalizeCodeUnitIDs(snap *TestSuiteSnapshot) {
	if snap == nil {
		return
	}
	for i := range snap.CodeUnits {
		cu := &snap.CodeUnits[i]
		if cu.UnitID != "" || cu.Path == "" || cu.Name == "" {
			continue
		}
		if cu.ParentName != "" {
			cu.UnitID = cu.Path + ":" + cu.ParentName + "." + cu.Name
			continue
		}
		cu.UnitID = cu.Path + ":" + cu.Name
	}
}

// sortSignals sorts a slice of signals into canonical order.
func sortSignals(signals []Signal) {
	sort.Slice(signals, func(i, j int) bool {
		a, b := signals[i], signals[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Location.File != b.Location.File {
			return a.Location.File < b.Location.File
		}
		if a.Location.Line != b.Location.Line {
			return a.Location.Line < b.Location.Line
		}
		return a.Explanation < b.Explanation
	})
}

func normalizeSnapshotMaps(snap *TestSuiteSnapshot) {
	if snap == nil {
		return
	}
	for path, owners := range snap.Ownership {
		if len(owners) <= 1 {
			continue
		}
		dup := append([]string(nil), owners...)
		sort.Strings(dup)
		snap.Ownership[path] = dedupeStrings(dup)
	}

	if rulesRaw, ok := snap.Policies["rules"]; ok {
		if rules, ok := rulesRaw.(map[string]any); ok {
			if frameworksRaw, ok := rules["disallow_frameworks"]; ok {
				if frameworks, ok := toStringSlice(frameworksRaw); ok {
					sort.Strings(frameworks)
					rules["disallow_frameworks"] = frameworks
				}
			}
		}
	}
}

func toStringSlice(v any) ([]string, bool) {
	switch arr := v.(type) {
	case []string:
		out := append([]string(nil), arr...)
		return out, true
	case []any:
		out := make([]string, 0, len(arr))
		for _, item := range arr {
			s, ok := item.(string)
			if !ok {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	default:
		return nil, false
	}
}

func dedupeStrings(items []string) []string {
	if len(items) < 2 {
		return items
	}
	out := items[:0]
	for i, item := range items {
		if i == 0 || item != items[i-1] {
			out = append(out, item)
		}
	}
	return out
}
