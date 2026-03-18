package analysis

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestResolveGoSymbolLinks_FunctionCalls(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Source file with exported functions.
	writeTempFile(t, root, "src/calc.go", `package calc

func Add(a, b int) int { return a + b }
func Subtract(a, b int) int { return a - b }
func Multiply(a, b int) int { return a * b }
func internalHelper() {}
`)

	// Test file that calls Add and Subtract but not Multiply.
	writeTempFile(t, root, "src/calc_test.go", `package calc

import "testing"

func TestAdd(t *testing.T) {
	result := Add(1, 2)
	if result != 3 {
		t.Errorf("Add(1, 2) = %d, want 3", result)
	}
}

func TestSubtract(t *testing.T) {
	result := Subtract(5, 3)
	if result != 2 {
		t.Errorf("Subtract(5, 3) = %d, want 2", result)
	}
}
`)

	candidates := []models.CodeUnit{
		{UnitID: "src/calc.go:Add", Name: "Add", Path: "src/calc.go", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/calc.go:Subtract", Name: "Subtract", Path: "src/calc.go", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/calc.go:Multiply", Name: "Multiply", Path: "src/calc.go", Kind: models.CodeUnitKindFunction, Exported: true},
	}

	links := resolveGoSymbolLinks(root, "src/calc_test.go", candidates)

	linkedIDs := map[string]float64{}
	for _, l := range links {
		linkedIDs[l.CodeUnitID] = l.Confidence
	}

	// Add and Subtract should be linked (directly called).
	if _, ok := linkedIDs["src/calc.go:Add"]; !ok {
		t.Error("expected Add to be linked — it's called in TestAdd")
	}
	if _, ok := linkedIDs["src/calc.go:Subtract"]; !ok {
		t.Error("expected Subtract to be linked — it's called in TestSubtract")
	}

	// Multiply should NOT be linked — it's never called.
	if _, ok := linkedIDs["src/calc.go:Multiply"]; ok {
		t.Error("Multiply should not be linked — it's never called in the test")
	}

	// Confidence should be 1.0 for AST-verified calls.
	if c := linkedIDs["src/calc.go:Add"]; c != 1.0 {
		t.Errorf("Add confidence: want 1.0, got %f", c)
	}
}

func TestResolveGoSymbolLinks_MethodCalls(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "src/svc_test.go", `package svc

import "testing"

func TestService(t *testing.T) {
	s := NewService()
	result := s.Process("input")
	_ = result
}
`)

	candidates := []models.CodeUnit{
		{UnitID: "src/svc.go:NewService", Name: "NewService", Path: "src/svc.go", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/svc.go:Service.Process", Name: "Process", Path: "src/svc.go", Kind: models.CodeUnitKindMethod, ParentName: "Service", Exported: true},
		{UnitID: "src/svc.go:Service.Cleanup", Name: "Cleanup", Path: "src/svc.go", Kind: models.CodeUnitKindMethod, ParentName: "Service", Exported: true},
	}

	links := resolveGoSymbolLinks(root, "src/svc_test.go", candidates)

	linkedIDs := map[string]bool{}
	for _, l := range links {
		linkedIDs[l.CodeUnitID] = true
	}

	if !linkedIDs["src/svc.go:NewService"] {
		t.Error("expected NewService to be linked")
	}
	if !linkedIDs["src/svc.go:Service.Process"] {
		t.Error("expected Service.Process to be linked")
	}
	if linkedIDs["src/svc.go:Service.Cleanup"] {
		t.Error("Cleanup should not be linked — it's never called")
	}
}

func TestResolveJSSymbolLinks_NamedImports(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "test/auth.test.ts", `
import { login, register } from '../src/auth';

describe('Auth', () => {
  it('should login', () => {
    const result = login('user', 'pass');
    expect(result).toBeDefined();
  });
});
`)

	candidates := []models.CodeUnit{
		{UnitID: "src/auth.ts:login", Name: "login", Path: "src/auth.ts", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/auth.ts:register", Name: "register", Path: "src/auth.ts", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/auth.ts:validateToken", Name: "validateToken", Path: "src/auth.ts", Kind: models.CodeUnitKindFunction, Exported: true},
	}

	links := resolveJSSymbolLinks(root, "test/auth.test.ts", candidates)

	linkedIDs := map[string]float64{}
	for _, l := range links {
		linkedIDs[l.CodeUnitID] = l.Confidence
	}

	// login: imported AND called → high confidence.
	if c, ok := linkedIDs["src/auth.ts:login"]; !ok {
		t.Error("expected login to be linked")
	} else if c < 0.9 {
		t.Errorf("login confidence: want >= 0.9, got %f", c)
	}

	// register: imported but NOT called → still linked (imported names are strong evidence).
	if _, ok := linkedIDs["src/auth.ts:register"]; !ok {
		t.Error("expected register to be linked — it's imported")
	}

	// validateToken: not imported, not called → not linked.
	if _, ok := linkedIDs["src/auth.ts:validateToken"]; ok {
		t.Error("validateToken should not be linked — it's not imported or called")
	}
}

func TestResolvePythonSymbolLinks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "tests/test_refund.py", `
from src.billing import calculate_refund, process_payment

def test_refund():
    result = calculate_refund(100, 0.1)
    assert result == 90

def test_payment():
    # process_payment is imported but not called in this test
    pass
`)

	candidates := []models.CodeUnit{
		{UnitID: "src/billing.py:calculate_refund", Name: "calculate_refund", Path: "src/billing.py", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/billing.py:process_payment", Name: "process_payment", Path: "src/billing.py", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/billing.py:apply_discount", Name: "apply_discount", Path: "src/billing.py", Kind: models.CodeUnitKindFunction, Exported: true},
	}

	links := resolvePythonSymbolLinks(root, "tests/test_refund.py", candidates)

	linkedIDs := map[string]float64{}
	for _, l := range links {
		linkedIDs[l.CodeUnitID] = l.Confidence
	}

	// calculate_refund: imported AND called.
	if _, ok := linkedIDs["src/billing.py:calculate_refund"]; !ok {
		t.Error("expected calculate_refund to be linked")
	}

	// process_payment: imported (visible in from...import).
	if _, ok := linkedIDs["src/billing.py:process_payment"]; !ok {
		t.Error("expected process_payment to be linked — it's imported")
	}

	// apply_discount: not imported, not referenced.
	if _, ok := linkedIDs["src/billing.py:apply_discount"]; ok {
		t.Error("apply_discount should not be linked")
	}
}

func TestPopulateSymbolLinks_SymbolLevel(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Source file with 3 exported functions.
	writeTempFile(t, root, "src/utils.ts", `
export function parseConfig(raw) { return JSON.parse(raw); }
export function formatOutput(data) { return JSON.stringify(data); }
export function validateInput(input) { return input != null; }
`)

	// Test file imports parseConfig and formatOutput, calls only parseConfig.
	writeTempFile(t, root, "test/utils.test.ts", `
import { parseConfig, formatOutput } from '../src/utils';

describe('Utils', () => {
  it('parses config', () => {
    const result = parseConfig('{"key": "value"}');
    expect(result.key).toBe('value');
  });
});
`)

	testFiles := []models.TestFile{
		{Path: "test/utils.test.ts", Framework: "jest"},
	}
	codeUnits := []models.CodeUnit{
		{UnitID: "src/utils.ts:parseConfig", Name: "parseConfig", Path: "src/utils.ts", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/utils.ts:formatOutput", Name: "formatOutput", Path: "src/utils.ts", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/utils.ts:validateInput", Name: "validateInput", Path: "src/utils.ts", Kind: models.CodeUnitKindFunction, Exported: true},
	}
	importGraph := &ImportGraph{
		TestImports: map[string]map[string]bool{
			"test/utils.test.ts": {"src/utils.ts": true},
		},
	}

	PopulateSymbolLinks(root, testFiles, codeUnits, importGraph)

	linked := testFiles[0].LinkedCodeUnits
	linkedSet := map[string]bool{}
	for _, id := range linked {
		linkedSet[id] = true
	}

	// parseConfig: imported + called → linked.
	if !linkedSet["src/utils.ts:parseConfig"] {
		t.Error("expected parseConfig to be linked")
	}

	// formatOutput: imported → linked (even though not called, it's explicitly imported).
	if !linkedSet["src/utils.ts:formatOutput"] {
		t.Error("expected formatOutput to be linked — it's imported")
	}

	// validateInput: NOT imported, NOT called → NOT linked.
	if linkedSet["src/utils.ts:validateInput"] {
		t.Error("validateInput should NOT be linked — it's not imported or referenced")
	}
}

func TestPopulateSymbolLinks_FallbackToFileLevel(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Test file with no parseable content (e.g., binary or unsupported format).
	writeTempFile(t, root, "test/mystery.test.rb", `
# Ruby test — not supported for symbol resolution
`)

	testFiles := []models.TestFile{
		{Path: "test/mystery.test.rb", Framework: "rspec"},
	}
	codeUnits := []models.CodeUnit{
		{UnitID: "src/mystery.rb:do_thing", Name: "do_thing", Path: "src/mystery.rb", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/mystery.rb:other_thing", Name: "other_thing", Path: "src/mystery.rb", Kind: models.CodeUnitKindFunction, Exported: true},
	}
	importGraph := &ImportGraph{
		TestImports: map[string]map[string]bool{
			"test/mystery.test.rb": {"src/mystery.rb": true},
		},
	}

	PopulateSymbolLinks(root, testFiles, codeUnits, importGraph)

	linked := testFiles[0].LinkedCodeUnits

	// Fallback: all units from imported file should be linked.
	if len(linked) != 2 {
		t.Errorf("expected 2 linked units (file-level fallback), got %d", len(linked))
	}
}

func TestProtectionGapReporting_SymbolGranularity(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Source file with 3 functions.
	writeTempFile(t, root, "src/billing.ts", `
export function calculateRefund(amount, rate) { return amount * (1 - rate); }
export function processPayment(amount) { /* ... */ }
export function applyDiscount(amount, pct) { return amount * (1 - pct); }
`)

	// Test that only calls calculateRefund.
	writeTempFile(t, root, "test/billing.test.ts", `
import { calculateRefund } from '../src/billing';

describe('Billing', () => {
  it('calculates refund', () => {
    expect(calculateRefund(100, 0.1)).toBe(90);
  });
});
`)

	testFiles := []models.TestFile{
		{Path: "test/billing.test.ts", Framework: "jest"},
	}
	codeUnits := []models.CodeUnit{
		{UnitID: "src/billing.ts:calculateRefund", Name: "calculateRefund", Path: "src/billing.ts", Kind: models.CodeUnitKindFunction, Exported: true, StartLine: 2},
		{UnitID: "src/billing.ts:processPayment", Name: "processPayment", Path: "src/billing.ts", Kind: models.CodeUnitKindFunction, Exported: true, StartLine: 3},
		{UnitID: "src/billing.ts:applyDiscount", Name: "applyDiscount", Path: "src/billing.ts", Kind: models.CodeUnitKindFunction, Exported: true, StartLine: 4},
	}
	importGraph := &ImportGraph{
		TestImports: map[string]map[string]bool{
			"test/billing.test.ts": {"src/billing.ts": true},
		},
	}

	PopulateSymbolLinks(root, testFiles, codeUnits, importGraph)

	linked := testFiles[0].LinkedCodeUnits
	linkedSet := map[string]bool{}
	for _, id := range linked {
		linkedSet[id] = true
	}

	// calculateRefund: imported + called → linked.
	if !linkedSet["src/billing.ts:calculateRefund"] {
		t.Error("calculateRefund should be linked")
	}

	// processPayment and applyDiscount: NOT imported → NOT linked.
	// This is the key improvement: at file-level, they would have been linked.
	if linkedSet["src/billing.ts:processPayment"] {
		t.Error("processPayment should NOT be linked at symbol level")
	}
	if linkedSet["src/billing.ts:applyDiscount"] {
		t.Error("applyDiscount should NOT be linked at symbol level")
	}
}

func TestContainsCallSite(t *testing.T) {
	t.Parallel()
	cases := []struct {
		src  string
		name string
		want bool
	}{
		{"result = Add(1, 2)", "Add", true},
		{"result = add(1, 2)", "Add", false},
		{"someAdd(1, 2)", "Add", false},
		{"x.Process(input)", "Process", true},
		{"no match here", "Foo", false},
		{"Foo()", "Foo", true},
	}
	for _, tc := range cases {
		got := containsCallSite(tc.src, tc.name)
		if got != tc.want {
			t.Errorf("containsCallSite(%q, %q) = %v, want %v", tc.src, tc.name, got, tc.want)
		}
	}
}

func TestContainsWordBoundary(t *testing.T) {
	t.Parallel()
	cases := []struct {
		src  string
		name string
		want bool
	}{
		{"import { login } from './auth'", "login", true},
		{"loginHandler()", "login", false},
		{"const x = login()", "login", true},
		{"no match", "foo", false},
	}
	for _, tc := range cases {
		got := containsWordBoundary(tc.src, tc.name)
		if got != tc.want {
			t.Errorf("containsWordBoundary(%q, %q) = %v, want %v", tc.src, tc.name, got, tc.want)
		}
	}
}

func TestExtractJSImportedNames(t *testing.T) {
	t.Parallel()
	src := `
import { login, register as signup } from '../src/auth';
import DefaultExport from '../src/main';
const { parseConfig } = require('./utils');
`
	names := extractJSImportedNames(src)

	expected := []string{"login", "signup", "DefaultExport", "parseConfig"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected imported name %q", name)
		}
	}
	// "register" should NOT appear (aliased to "signup").
	if names["register"] {
		t.Error("register should not appear — it's aliased to signup")
	}
}

func TestExtractPythonImportedNames(t *testing.T) {
	t.Parallel()
	src := `
from src.billing import calculate_refund, process_payment
from src.auth import login as authenticate
import os
`
	names := extractPythonImportedNames(src)

	expected := []string{"calculate_refund", "process_payment", "authenticate"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected imported name %q", name)
		}
	}
	// "login" should NOT appear (aliased to "authenticate").
	if names["login"] {
		t.Error("login should not appear — it's aliased to authenticate")
	}
}

func TestResolveSymbolLinks_EmptyInputs(t *testing.T) {
	t.Parallel()

	// nil import graph.
	links := ResolveSymbolLinks("/tmp", nil, nil, nil)
	if len(links) != 0 {
		t.Errorf("expected 0 links for nil graph, got %d", len(links))
	}

	// Empty test imports.
	links = ResolveSymbolLinks("/tmp", nil, []models.CodeUnit{{Name: "foo"}}, &ImportGraph{
		TestImports: map[string]map[string]bool{},
	})
	if len(links) != 0 {
		t.Errorf("expected 0 links for empty imports, got %d", len(links))
	}
}

func TestPopulateSymbolLinks_Deterministic(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "test/a.test.ts", `
import { foo, bar, baz } from '../src/mod';
foo(); bar(); baz();
`)

	testFiles := []models.TestFile{
		{Path: "test/a.test.ts", Framework: "jest"},
	}
	codeUnits := []models.CodeUnit{
		{UnitID: "src/mod.ts:foo", Name: "foo", Path: "src/mod.ts", Kind: models.CodeUnitKindFunction},
		{UnitID: "src/mod.ts:bar", Name: "bar", Path: "src/mod.ts", Kind: models.CodeUnitKindFunction},
		{UnitID: "src/mod.ts:baz", Name: "baz", Path: "src/mod.ts", Kind: models.CodeUnitKindFunction},
	}
	importGraph := &ImportGraph{
		TestImports: map[string]map[string]bool{
			"test/a.test.ts": {"src/mod.ts": true},
		},
	}

	// Run twice to verify determinism.
	PopulateSymbolLinks(root, testFiles, codeUnits, importGraph)
	first := make([]string, len(testFiles[0].LinkedCodeUnits))
	copy(first, testFiles[0].LinkedCodeUnits)

	testFiles[0].LinkedCodeUnits = nil
	PopulateSymbolLinks(root, testFiles, codeUnits, importGraph)
	second := testFiles[0].LinkedCodeUnits

	sort.Strings(first)
	sort.Strings(second)

	if len(first) != len(second) {
		t.Fatalf("non-deterministic: first=%d, second=%d", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("non-deterministic at index %d: %s != %s", i, first[i], second[i])
		}
	}
}

func writeTempFileForTest(t *testing.T, root, relPath, content string) {
	t.Helper()
	absPath := filepath.Join(root, relPath)
	os.MkdirAll(filepath.Dir(absPath), 0o755)
	os.WriteFile(absPath, []byte(content), 0o644)
}
