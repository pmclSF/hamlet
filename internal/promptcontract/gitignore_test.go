package promptcontract

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAnalyzeRepo_HonoursGitignore locks the walker's .gitignore contract:
// prompts, schemas, and drift under an ignored directory contribute nothing
// to the inventory, while files that are merely untracked (present on disk,
// not ignored) are still analyzed.
func TestAnalyzeRepo_HonoursGitignore(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write(".gitignore", "local/\n")

	// Ignored directory: a schema + a drifting prompt that must not surface.
	write("local/models.py", "from pydantic import BaseModel\n\n"+
		"class Ignored(BaseModel):\n    name: str\n")
	write("local/prompt.py", "import openai\nfrom models import Ignored\n\n"+
		"def build(user: Ignored) -> str:\n"+
		"    return f\"\"\"Hello {user.user_id}.\"\"\"\n")

	// Untracked but not ignored: a consistent schema + prompt pair.
	write("models.py", "from pydantic import BaseModel\n\n"+
		"class UserProfile(BaseModel):\n    user_id: str\n")
	write("prompt.py", "import openai\nfrom models import UserProfile\n\n"+
		"def build(user: UserProfile) -> str:\n"+
		"    return f\"\"\"Hi {user.user_id}.\"\"\"\n")

	inv, drift, err := AnalyzeRepo(root)
	if err != nil {
		t.Fatalf("AnalyzeRepo: %v", err)
	}
	if inv.Prompts != 1 || inv.PromptFiles != 1 || inv.Schemas != 1 {
		t.Errorf("inventory counted gitignored surfaces: %+v (want 1 prompt / 1 file / 1 schema)", inv)
	}
	if len(drift) != 0 {
		t.Errorf("drift from a gitignored directory must not surface, got %+v", drift)
	}
}
