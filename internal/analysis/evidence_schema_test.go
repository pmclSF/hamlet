package analysis

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- Evidence completeness: every surface must carry tier + confidence ---

func TestEvidenceSchema_AllCodeSurfacesHaveTierAndConfidence(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "src/app.ts", `
import express from 'express';
const app = express();
app.get('/api/users', (req, res) => { res.json([]); });
export function loginHandler(req, res) {}
export class UserService {}
`)

	surfaces := InferCodeSurfaces(root, nil)
	violations := models.ValidateSurfaceTiers(surfaces)
	if len(violations) > 0 {
		t.Errorf("surfaces with invalid evidence: %v", violations)
	}
}

func TestEvidenceSchema_AIContextSurfacesHaveReasonWithDetectorID(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "src/chat.ts", `
const messages = [
  { role: "system", content: "You are a helpful assistant." },
  { role: "user", content: userInput },
];
`)

	surfaces := InferAIContextSurfaces(root, nil, nil)
	for _, s := range surfaces {
		if s.Reason == "" {
			t.Errorf("AI context surface %q missing Reason", s.Name)
			continue
		}
		if !strings.HasPrefix(s.Reason, "[") {
			t.Errorf("AI context surface %q Reason should start with [detectorID], got: %s", s.Name, s.Reason)
		}
	}
}

func TestEvidenceSchema_FixtureSurfacesHaveReason(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "test/helpers.ts", `
beforeEach(() => { setup(); });
afterAll(() => { teardown(); });
export function createUser(data) { return { ...data }; }
const mockApi = jest.fn();
`)

	testFiles := []models.TestFile{
		{Path: "test/helpers.ts", Framework: "jest"},
	}

	fixtures := ExtractFixtures(root, testFiles)
	if len(fixtures) == 0 {
		t.Fatal("expected fixtures to be detected")
	}

	for _, f := range fixtures {
		if f.DetectionTier == "" {
			t.Errorf("fixture %q missing DetectionTier", f.Name)
		}
		if f.Confidence == 0 {
			t.Errorf("fixture %q has zero Confidence", f.Name)
		}
		if f.Reason == "" {
			t.Errorf("fixture %q missing Reason", f.Name)
		}
		if !strings.HasPrefix(f.Reason, "[") {
			t.Errorf("fixture %q Reason should start with [detectorID], got: %s", f.Name, f.Reason)
		}
	}
}

// --- Evidence() bridge method tests ---

func TestEvidenceSchema_CodeSurfaceEvidenceBridge(t *testing.T) {
	t.Parallel()
	cs := models.CodeSurface{
		SurfaceID:     "surface:src/api.ts:login",
		Name:          "login",
		Path:          "src/api.ts",
		Kind:          models.SurfaceFunction,
		Line:          10,
		DetectionTier: models.TierPattern,
		Confidence:    0.90,
		Reason:        "[pattern:export-name] exported function 'login'",
	}

	ev := cs.Evidence()
	if ev.Tier != models.TierPattern {
		t.Errorf("tier: want pattern, got %s", ev.Tier)
	}
	if ev.Confidence != 0.90 {
		t.Errorf("confidence: want 0.90, got %f", ev.Confidence)
	}
	if ev.FilePath != "src/api.ts" {
		t.Errorf("filePath: want src/api.ts, got %s", ev.FilePath)
	}
	if ev.Symbol != "login" {
		t.Errorf("symbol: want login, got %s", ev.Symbol)
	}
	if ev.Line != 10 {
		t.Errorf("line: want 10, got %d", ev.Line)
	}
	if ev.Reason != cs.Reason {
		t.Errorf("reason not preserved")
	}
}

func TestEvidenceSchema_FixtureSurfaceEvidenceBridge(t *testing.T) {
	t.Parallel()
	fs := models.FixtureSurface{
		FixtureID:     "fixture:test/h.ts:beforeEach",
		Name:          "beforeEach",
		Path:          "test/h.ts",
		Line:          5,
		DetectionTier: models.TierPattern,
		Confidence:    0.95,
		Reason:        "[pattern:fixture-lifecycle-hook] setup_hook 'beforeEach' (scope: test)",
	}

	ev := fs.Evidence()
	if ev.Tier != models.TierPattern {
		t.Errorf("tier: want pattern, got %s", ev.Tier)
	}
	if ev.Symbol != "beforeEach" {
		t.Errorf("symbol: want beforeEach, got %s", ev.Symbol)
	}
	if ev.Reason == "" {
		t.Error("reason should be non-empty")
	}
}

func TestEvidenceSchema_RAGPipelineSurfaceEvidenceBridge(t *testing.T) {
	t.Parallel()
	rs := models.RAGPipelineSurface{
		ComponentID:   "rag:src/s.py:retriever:retriever",
		Name:          "retriever_config",
		Path:          "src/s.py",
		Line:          20,
		DetectionTier: models.TierSemantic,
		Confidence:    0.92,
		Reason:        "[semantic:rag-retriever] .as_retriever() construction (top_k=5)",
	}

	ev := rs.Evidence()
	if ev.Tier != models.TierSemantic {
		t.Errorf("tier: want semantic, got %s", ev.Tier)
	}
	if ev.Confidence != 0.92 {
		t.Errorf("confidence: want 0.92, got %f", ev.Confidence)
	}
}

// --- FormatReason ---

func TestFormatReason(t *testing.T) {
	t.Parallel()
	r := models.FormatReason("structural:go-export", "exported function 'Hello'")
	if r != "[structural:go-export] exported function 'Hello'" {
		t.Errorf("unexpected format: %s", r)
	}

	r2 := models.FormatReason("", "plain description")
	if r2 != "plain description" {
		t.Errorf("empty detectorID should return plain description, got: %s", r2)
	}
}

// --- Reason format consistency across all parsers ---

func TestEvidenceSchema_ReasonFormatConsistency(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create a source file with multiple surface types to trigger various detectors.
	writeTempFile(t, root, "src/ai.ts", `
import { SystemMessage, HumanMessage } from "@langchain/core/messages";

const messages = [
  { role: "system", content: "You are a helpful assistant." },
  { role: "user", content: userInput },
];

const systemPrompt = "You are a helpful coding assistant. Your role is to answer questions accurately.";

export function buildPrompt(query) {
  return [
    { role: "system", content: systemPrompt },
    { role: "user", content: query },
  ];
}
`)

	surfaces := InferAIContextSurfaces(root, nil, nil)
	for _, s := range surfaces {
		if s.Reason == "" {
			continue // Some name-based surfaces don't have Reason (e.g., standard exports)
		}
		// Every populated Reason should follow [detectorID] format.
		if !strings.HasPrefix(s.Reason, "[") {
			t.Errorf("surface %q has non-standard Reason format: %s", s.Name, s.Reason)
		}
		// The detectorID inside brackets should contain a colon (tier:name format).
		if bracketEnd := strings.Index(s.Reason, "]"); bracketEnd > 1 {
			detectorID := s.Reason[1:bracketEnd]
			if !strings.Contains(detectorID, ":") {
				t.Errorf("surface %q Reason has malformed detectorID: %s", s.Name, detectorID)
			}
			tier := models.TierFromDetectorID(detectorID)
			if !models.IsValidTier(tier) {
				t.Errorf("surface %q Reason detectorID %q has invalid tier prefix", s.Name, detectorID)
			}
		}
	}
}

// --- JSON serialization ---

func TestEvidenceSchema_CodeSurfaceJSONRoundTrip(t *testing.T) {
	t.Parallel()
	cs := models.CodeSurface{
		SurfaceID:     "surface:src/api.ts:login",
		Name:          "login",
		Path:          "src/api.ts",
		Kind:          models.SurfaceFunction,
		Language:      "js",
		Line:          10,
		Exported:      true,
		DetectionTier: models.TierPattern,
		Confidence:    0.90,
		Reason:        "[pattern:export-name] exported function 'login'",
	}

	data, err := json.Marshal(cs)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify all evidence fields are present in JSON.
	jsonStr := string(data)
	for _, field := range []string{
		`"detectionTier"`, `"confidence"`, `"reason"`,
		`"surfaceId"`, `"name"`, `"path"`, `"line"`,
	} {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON missing field %s: %s", field, jsonStr)
		}
	}

	// Round-trip.
	var decoded models.CodeSurface
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.DetectionTier != cs.DetectionTier {
		t.Errorf("round-trip tier: want %s, got %s", cs.DetectionTier, decoded.DetectionTier)
	}
	if decoded.Confidence != cs.Confidence {
		t.Errorf("round-trip confidence: want %f, got %f", cs.Confidence, decoded.Confidence)
	}
	if decoded.Reason != cs.Reason {
		t.Errorf("round-trip reason: want %s, got %s", cs.Reason, decoded.Reason)
	}
}

func TestEvidenceSchema_FixtureSurfaceJSONIncludesReason(t *testing.T) {
	t.Parallel()
	fs := models.FixtureSurface{
		FixtureID:     "fixture:test/h.ts:beforeEach",
		Name:          "beforeEach",
		Path:          "test/h.ts",
		Kind:          models.FixtureSetupHook,
		Line:          5,
		DetectionTier: models.TierPattern,
		Confidence:    0.95,
		Reason:        "[pattern:fixture-lifecycle-hook] setup_hook 'beforeEach'",
	}

	data, err := json.Marshal(fs)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"reason"`) {
		t.Errorf("fixture JSON missing reason field: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"detectionTier"`) {
		t.Errorf("fixture JSON missing detectionTier field: %s", jsonStr)
	}
}

func TestEvidenceSchema_DetectionEvidenceJSONRoundTrip(t *testing.T) {
	t.Parallel()
	ev := models.DetectionEvidence{
		DetectorID: models.DetectorASTMessageArray,
		Tier:       models.TierStructural,
		Confidence: 0.97,
		FilePath:   "src/chat.ts",
		Symbol:     "messages",
		Line:       5,
		Reason:     "[structural:ast-message-array] AST-verified message array",
	}

	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded models.DetectionEvidence
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.DetectorID != ev.DetectorID {
		t.Errorf("detectorId: want %s, got %s", ev.DetectorID, decoded.DetectorID)
	}
	if decoded.Tier != ev.Tier {
		t.Errorf("tier: want %s, got %s", ev.Tier, decoded.Tier)
	}
	if decoded.Confidence != ev.Confidence {
		t.Errorf("confidence: want %f, got %f", ev.Confidence, decoded.Confidence)
	}
}

// --- Deterministic ordering ---

func TestEvidenceSchema_SurfaceEvidenceDeterministic(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/svc.ts", `
export function handleAuth(req, res) {}
export function processPayment(req, res) {}
export class OrderService {}
`)

	s1 := InferCodeSurfaces(root, nil)
	s2 := InferCodeSurfaces(root, nil)

	if len(s1) != len(s2) {
		t.Fatalf("non-deterministic count: %d vs %d", len(s1), len(s2))
	}

	for i := range s1 {
		if s1[i].DetectionTier != s2[i].DetectionTier {
			t.Errorf("[%d] tier differs: %s vs %s", i, s1[i].DetectionTier, s2[i].DetectionTier)
		}
		if s1[i].Confidence != s2[i].Confidence {
			t.Errorf("[%d] confidence differs: %f vs %f", i, s1[i].Confidence, s2[i].Confidence)
		}
	}
}

// --- Framework detection evidence ---

func TestEvidenceSchema_FrameworkDetectorIDIsPopulated(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/auth.test.ts", `
import { describe, it, expect } from 'vitest';
describe('auth', () => { it('works', () => { expect(true).toBe(true); }); });
`)

	testFiles, err := discoverTestFiles(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, tf := range testFiles {
		if tf.FrameworkSource == "" {
			continue // Skip files with no framework detection
		}
		if tf.FrameworkDetectorID == "" {
			t.Errorf("test file %q has FrameworkSource=%s but no FrameworkDetectorID",
				tf.Path, tf.FrameworkSource)
		}
		// DetectorID should have valid tier prefix.
		tier := models.TierFromDetectorID(tf.FrameworkDetectorID)
		if tier != "" && !models.IsValidTier(tier) {
			t.Errorf("test file %q has invalid tier in FrameworkDetectorID: %s",
				tf.Path, tf.FrameworkDetectorID)
		}
	}
}
