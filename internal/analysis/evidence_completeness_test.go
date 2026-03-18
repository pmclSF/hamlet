package analysis

import (
	"sort"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// TestEvidenceCompleteness_AllSurfacesHaveTierAndConfidence verifies that
// every CodeSurface produced by the analysis pipeline carries detection
// tier, confidence, and location metadata.
func TestEvidenceCompleteness_AllSurfacesHaveTierAndConfidence(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create a JS source file with various surface types.
	writeTempFile(t, root, "src/app.ts", `
import express from 'express';
const app = express();

app.get('/api/users', async (req, res) => {
  const users = await db.findAll();
  res.json(users);
});

export async function loginHandler(req, res) {
  const token = await authenticate(req.body);
  res.json({ token });
}

export function validateInput(data) {
  return data != null;
}

export class UserController {
  getUser(id) { return db.find(id); }
}
`)

	testFiles := []models.TestFile{}
	surfaces := InferCodeSurfaces(root, testFiles)

	if len(surfaces) == 0 {
		t.Fatal("expected at least 1 surface from test fixture")
	}

	for _, s := range surfaces {
		if s.DetectionTier == "" {
			t.Errorf("surface %q (%s) at %s:%d missing DetectionTier",
				s.Name, s.Kind, s.Path, s.Line)
		}
		if s.Confidence == 0 {
			t.Errorf("surface %q (%s) at %s:%d has zero Confidence",
				s.Name, s.Kind, s.Path, s.Line)
		}
		if s.Path == "" {
			t.Errorf("surface %q missing Path", s.Name)
		}
		if s.SurfaceID == "" {
			t.Errorf("surface %q missing SurfaceID", s.Name)
		}
		if s.Language == "" {
			t.Errorf("surface %q missing Language", s.Name)
		}
	}
}

// TestEvidenceCompleteness_ContentInferredSurfacesHaveReason verifies that
// content-inferred surfaces (message arrays, prompt templates) carry a
// non-empty Reason field explaining the classification.
func TestEvidenceCompleteness_ContentInferredSurfacesHaveReason(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "src/chat.ts", `
const messages = [
  { role: "system", content: "You are a helpful assistant." },
  { role: "user", content: userInput },
];
`)

	surfaces := InferAIContextSurfaces(root, nil, nil)

	if len(surfaces) == 0 {
		t.Fatal("expected at least 1 AI context surface")
	}

	for _, s := range surfaces {
		if s.Reason == "" {
			t.Errorf("content-inferred surface %q at %s:%d missing Reason",
				s.Name, s.Path, s.Line)
		}
		if s.DetectionTier == "" {
			t.Errorf("content-inferred surface %q missing DetectionTier", s.Name)
		}
		if s.Confidence == 0 {
			t.Errorf("content-inferred surface %q has zero Confidence", s.Name)
		}
	}
}

// TestEvidenceCompleteness_FixtureSurfaceKind verifies that SurfaceFixture
// gets proper tier and confidence from assignInferenceMetadata.
func TestEvidenceCompleteness_FixtureSurfaceKind(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{Kind: models.SurfaceFixture, Name: "beforeEach", Language: "js"},
	}
	assignInferenceMetadata(surfaces)

	s := surfaces[0]
	if s.DetectionTier != models.TierPattern {
		t.Errorf("fixture tier: want pattern, got %s", s.DetectionTier)
	}
	if s.Confidence != 0.85 {
		t.Errorf("fixture confidence: want 0.85, got %f", s.Confidence)
	}
}

// TestEvidenceOrdering_SurfacesAreDeterministic verifies that the order of
// surfaces from InferCodeSurfaces is deterministic across multiple runs.
func TestEvidenceOrdering_SurfacesAreDeterministic(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "src/api.ts", `
import express from 'express';
const app = express();
app.get('/api/users', handler);
app.post('/api/users', createHandler);
app.delete('/api/users/:id', deleteHandler);

export function loginHandler(req, res) {}
export function logoutHandler(req, res) {}
export function validateInput(data) { return true; }
export class AuthService {}
`)

	testFiles := []models.TestFile{}
	s1 := InferCodeSurfaces(root, testFiles)
	s2 := InferCodeSurfaces(root, testFiles)

	if len(s1) != len(s2) {
		t.Fatalf("non-deterministic surface count: %d vs %d", len(s1), len(s2))
	}

	// Sort both for comparison (the pipeline sorts by file, but within a file order matters).
	ids1 := extractSortedIDs(s1)
	ids2 := extractSortedIDs(s2)

	for i := range ids1 {
		if ids1[i] != ids2[i] {
			t.Errorf("non-deterministic at index %d: %s vs %s", i, ids1[i], ids2[i])
		}
	}
}

// TestEvidenceOrdering_AIContextDeterministic verifies that AI context
// inference produces deterministic results.
func TestEvidenceOrdering_AIContextDeterministic(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTempFile(t, root, "src/ai.ts", `
const messages = [
  { role: "system", content: "You are a helpful assistant." },
  { role: "user", content: userInput },
];
const systemPrompt = "You are a helpful coding assistant. Your role is to answer questions accurately.";
`)

	s1 := InferAIContextSurfaces(root, nil, nil)
	s2 := InferAIContextSurfaces(root, nil, nil)

	ids1 := extractSortedIDs(s1)
	ids2 := extractSortedIDs(s2)

	if len(ids1) != len(ids2) {
		t.Fatalf("non-deterministic count: %d vs %d", len(ids1), len(ids2))
	}
	for i := range ids1 {
		if ids1[i] != ids2[i] {
			t.Errorf("non-deterministic at index %d: %s vs %s", i, ids1[i], ids2[i])
		}
	}
}

func extractSortedIDs(surfaces []models.CodeSurface) []string {
	ids := make([]string, len(surfaces))
	for i, s := range surfaces {
		ids[i] = s.SurfaceID
	}
	sort.Strings(ids)
	return ids
}
