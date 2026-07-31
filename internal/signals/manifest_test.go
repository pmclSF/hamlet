package signals

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// TestManifest_MatchesSignalTypes is the drift gate for 0.1.2.
//
// It parses signal_types.go to discover every `Signal*` constant and asserts
// that allSignalManifest contains exactly one entry per constant, with the
// ConstName field set correctly. The test is intentionally bidirectional: a
// constant added without a manifest entry fails, and a manifest entry whose
// ConstName doesn't refer to a real constant also fails.
//
// Failures here indicate one of:
//   - a new SignalXxx constant was added but the manifest entry is missing
//   - a manifest entry references a constant that has been deleted or renamed
//   - the manifest entry's Type field disagrees with the constant's value
func TestManifest_MatchesSignalTypes(t *testing.T) {
	t.Parallel()

	// Locate signal_types.go alongside this test source.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)
	target := filepath.Join(dir, "signal_types.go")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, target, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse signal_types.go: %v", err)
	}

	// Collect every `SignalXxx models.SignalType = "..."` constant.
	type constInfo struct {
		Name  string
		Value models.SignalType
	}
	var consts []constInfo
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if !strings.HasPrefix(name.Name, "Signal") {
					continue
				}
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				value := strings.Trim(lit.Value, `"`)
				consts = append(consts, constInfo{Name: name.Name, Value: models.SignalType(value)})
			}
		}
	}

	if len(consts) == 0 {
		t.Fatalf("found zero Signal* constants in %s; parser likely broken", target)
	}

	// Build manifest indexes.
	byConst := make(map[string]ManifestEntry, len(allSignalManifest))
	byType := make(map[models.SignalType]ManifestEntry, len(allSignalManifest))
	dupConst := map[string]bool{}
	dupType := map[models.SignalType]bool{}
	for _, e := range allSignalManifest {
		if _, exists := byConst[e.ConstName]; exists {
			dupConst[e.ConstName] = true
		}
		byConst[e.ConstName] = e
		if _, exists := byType[e.Type]; exists {
			dupType[e.Type] = true
		}
		byType[e.Type] = e
	}
	for c := range dupConst {
		t.Errorf("manifest has duplicate ConstName %q", c)
	}
	for typ := range dupType {
		t.Errorf("manifest has duplicate Type %q", typ)
	}

	// Constants that have no manifest entry, or whose manifest Type differs.
	for _, c := range consts {
		entry, ok := byConst[c.Name]
		if !ok {
			t.Errorf("constant %s (%q) is missing from allSignalManifest", c.Name, c.Value)
			continue
		}
		if entry.Type != c.Value {
			t.Errorf(
				"manifest entry for %s has Type %q, want %q",
				c.Name, entry.Type, c.Value,
			)
		}
	}

	// Manifest entries that reference a non-existent constant.
	knownConsts := map[string]bool{}
	for _, c := range consts {
		knownConsts[c.Name] = true
	}
	for _, e := range allSignalManifest {
		if !knownConsts[e.ConstName] {
			t.Errorf(
				"manifest entry for type %q references constant %q which doesn't exist in signal_types.go",
				e.Type, e.ConstName,
			)
		}
	}

	// Sanity: the count should match.
	if len(consts) != len(allSignalManifest) {
		t.Errorf(
			"signal_types.go has %d Signal* constants but manifest has %d entries",
			len(consts), len(allSignalManifest),
		)
	}
}

// TestManifest_RuleIDsUnique guards against accidentally reusing a rule ID
// when copy-pasting an entry. Rule IDs are part of the public contract
// (SARIF emission, doc cross-references) so collisions are a release blocker.
func TestManifest_RuleIDsUnique(t *testing.T) {
	t.Parallel()

	seen := map[string]models.SignalType{}
	for _, e := range allSignalManifest {
		if e.RuleID == "" {
			t.Errorf("manifest entry %q has empty RuleID", e.Type)
			continue
		}
		if dup, exists := seen[e.RuleID]; exists {
			t.Errorf(
				"manifest entries %q and %q share RuleID %q",
				e.Type, dup, e.RuleID,
			)
			continue
		}
		seen[e.RuleID] = e.Type
	}
}

// TestManifest_PlannedHavePromotionPlan ensures every experimental and
// planned entry documents what it takes to promote it. Stable entries don't
// need this and may leave the field blank.
func TestManifest_PlannedHavePromotionPlan(t *testing.T) {
	t.Parallel()

	for _, e := range allSignalManifest {
		switch e.Status {
		case StatusPlanned, StatusExperimental:
			if strings.TrimSpace(e.PromotionPlan) == "" {
				t.Errorf(
					"entry %q has status %q but no PromotionPlan; document the path to stable",
					e.Type, e.Status,
				)
			}
		}
	}
}

// TestManifest_RegistryConsistent verifies that every entry in the legacy
// Registry map (registry.go) has a matching manifest entry. The Registry
// will be regenerated from the manifest in 0.2; for 0.1.2 we just guard
// against drift between the two while they coexist.
func TestManifest_RegistryConsistent(t *testing.T) {
	t.Parallel()

	for typ, def := range Registry {
		entry, ok := ManifestByType(typ)
		if !ok {
			t.Errorf("Registry has type %q but manifest does not", typ)
			continue
		}
		if entry.Domain != def.Category {
			t.Errorf(
				"Registry vs manifest disagree on category for %q: registry=%q manifest=%q",
				typ, def.Category, entry.Domain,
			)
		}
	}
}

// TestManifest_CatalogBidirectional verifies that every type in
// models.SignalCatalog has a manifest entry, and every manifest entry has
// a catalog entry. SignalCatalog already lists all 56 types but lacks
// status/severity/confidence/rule-ID metadata; the manifest is the
// superset. They merge in 0.2.
func TestManifest_CatalogBidirectional(t *testing.T) {
	t.Parallel()

	for typ := range models.SignalCatalog {
		if _, ok := ManifestByType(typ); !ok {
			t.Errorf("models.SignalCatalog has type %q but manifest does not", typ)
		}
	}
	for _, e := range allSignalManifest {
		if _, ok := models.SignalCatalog[e.Type]; !ok {
			t.Errorf("manifest has type %q but models.SignalCatalog does not", e.Type)
		}
	}
}

// TestTitleForType covers the shared title lookup used by user-facing
// renderers: manifest title when an entry exists, raw type string as
// the fallback.
func TestTitleForType(t *testing.T) {
	t.Parallel()

	if got := TitleForType(SignalWeakAssertion); got != "Weak Assertion" {
		t.Errorf("TitleForType(SignalWeakAssertion) = %q, want %q", got, "Weak Assertion")
	}
	if got := TitleForType(models.SignalType("noSuchSignal")); got != "noSuchSignal" {
		t.Errorf("TitleForType fallback = %q, want raw type string", got)
	}
}
