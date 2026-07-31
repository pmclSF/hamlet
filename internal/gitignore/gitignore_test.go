package gitignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatcher_BasicPatterns(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	body := []byte(`# comment line
node_modules/
*.log
/build
!build/keep.txt
dist
src/generated/
`)
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), body, 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	m := Load(root)
	if m == nil || len(m.rules) == 0 {
		t.Fatalf("expected gitignore rules to load")
	}

	cases := []struct {
		path  string
		isDir bool
		want  bool
	}{
		// Floating directory pattern
		{"node_modules", true, true},
		{"frontend/node_modules", true, true},
		// Files inside the floating dir match via the segment walk.
		{"node_modules/lodash/index.js", false, true},
		// Wildcard match anywhere
		{"foo.log", false, true},
		{"deep/nested/bar.log", false, true},
		// Anchored path
		{"build", true, true},
		{"build/output.txt", false, true},
		// Anchored path negation
		{"build/keep.txt", false, false},
		// Floating path matches at any depth
		{"dist", true, true},
		{"packages/foo/dist", true, true},
		// Nested anchored
		{"src/generated", true, true},
		{"src/generated/proto.pb.go", false, true},
		// Unrelated
		{"src/auth.ts", false, false},
		{"README.md", false, false},
	}

	for _, tc := range cases {
		got := m.Match(tc.path, tc.isDir)
		if got != tc.want {
			t.Errorf("Match(%q, isDir=%v) = %v, want %v", tc.path, tc.isDir, got, tc.want)
		}
	}
}

func TestMatcher_MissingFile(t *testing.T) {
	t.Parallel()

	m := Load(t.TempDir())
	if m == nil {
		t.Fatalf("expected non-nil matcher when .gitignore missing")
	}
	if m.Match("anything", false) {
		t.Errorf("expected no matches when .gitignore missing")
	}
}

func TestMatcher_NilReceiver(t *testing.T) {
	t.Parallel()

	var m *Matcher
	if m.Match("anything", false) {
		t.Errorf("nil matcher must match nothing")
	}
}
