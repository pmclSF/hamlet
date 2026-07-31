package aidetect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetect_HonoursGitignore locks the source walker's .gitignore contract:
// content under an ignored directory contributes no surfaces, while a file
// that is merely untracked (present on disk, not ignored) is still scanned.
func TestDetect_HonoursGitignore(t *testing.T) {
	t.Parallel()
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

	src := "import openai\nsystem_prompt = 'hi'\nopenai.ChatCompletion.create(model='gpt-4')\n"
	write(".gitignore", "local/\n")
	// Ignored directory: a full AI surface that must not be counted.
	write("local/agent.py", src)
	// Untracked but not ignored: must still be seen.
	write("src/agent.py", src)

	res := Detect(root)

	for _, p := range append(append([]string{}, res.PromptFiles...), res.ModelFiles...) {
		if strings.HasPrefix(filepath.ToSlash(p), "local/") {
			t.Errorf("gitignored path leaked into scan results: %s", p)
		}
	}
	if !containsPath(res.PromptFiles, "src/agent.py") {
		t.Errorf("untracked non-ignored prompt file missing from PromptFiles: %v", res.PromptFiles)
	}
	if !containsPath(res.ModelFiles, "src/agent.py") {
		t.Errorf("untracked non-ignored model file missing from ModelFiles: %v", res.ModelFiles)
	}
}

// TestWalkRepoForConfigs_HonoursGitignore covers both rule shapes on the
// config walker: an ignored directory is pruned, and a file-level pattern
// excludes an individual file.
func TestWalkRepoForConfigs_HonoursGitignore(t *testing.T) {
	t.Parallel()
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

	write(".gitignore", "scratch/\nsecret.yaml\n")
	write("scratch/promptfoo.yaml", "providers: []\n")
	write("secret.yaml", "providers: []\n")
	write("evals/promptfoo.yaml", "providers: []\n")

	got := walkRepoForConfigs(root, scanOpts{extensions: map[string]bool{".yaml": true}})
	if len(got) != 1 || got[0] != "evals/promptfoo.yaml" {
		t.Errorf("walkRepoForConfigs = %v, want [evals/promptfoo.yaml]", got)
	}
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if filepath.ToSlash(p) == want {
			return true
		}
	}
	return false
}
