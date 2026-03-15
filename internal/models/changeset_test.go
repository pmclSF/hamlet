package models

import (
	"encoding/json"
	"testing"
)

func TestChangeSet_SourceFiles(t *testing.T) {
	t.Parallel()
	cs := &ChangeSet{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js", IsTestFile: false},
			{Path: "test/auth.test.js", IsTestFile: true},
			{Path: "src/db.js", IsTestFile: false},
		},
	}

	src := cs.SourceFiles()
	if len(src) != 2 {
		t.Fatalf("expected 2 source files, got %d", len(src))
	}
	if src[0].Path != "src/auth.js" {
		t.Errorf("expected src/auth.js, got %s", src[0].Path)
	}
}

func TestChangeSet_TestFiles(t *testing.T) {
	t.Parallel()
	cs := &ChangeSet{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js", IsTestFile: false},
			{Path: "test/auth.test.js", IsTestFile: true},
		},
	}

	tests := cs.TestFiles()
	if len(tests) != 1 {
		t.Fatalf("expected 1 test file, got %d", len(tests))
	}
	if tests[0].Path != "test/auth.test.js" {
		t.Errorf("expected test/auth.test.js, got %s", tests[0].Path)
	}
}

func TestChangeSet_FileCount(t *testing.T) {
	t.Parallel()
	cs := &ChangeSet{
		ChangedFiles: []ChangedFile{
			{Path: "a.js"},
			{Path: "b.js"},
			{Path: "c.js"},
		},
	}
	if cs.FileCount() != 3 {
		t.Errorf("expected 3, got %d", cs.FileCount())
	}
}

func TestChangeSet_HasFile(t *testing.T) {
	t.Parallel()
	cs := &ChangeSet{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js"},
			{Path: "src/db.js"},
		},
	}

	if !cs.HasFile("src/auth.js") {
		t.Error("expected HasFile to return true for src/auth.js")
	}
	if cs.HasFile("src/missing.js") {
		t.Error("expected HasFile to return false for src/missing.js")
	}
}

func TestChangeSet_MarshalJSON(t *testing.T) {
	t.Parallel()
	cs := &ChangeSet{
		Source: "explicit",
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js", ChangeKind: ChangeModified},
		},
		ChangedPackages: []string{"src/auth"},
	}

	data, err := json.Marshal(cs)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var decoded ChangeSet
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if decoded.Source != "explicit" {
		t.Errorf("expected source 'explicit', got %q", decoded.Source)
	}
	if len(decoded.ChangedFiles) != 1 {
		t.Fatalf("expected 1 changed file, got %d", len(decoded.ChangedFiles))
	}
	if decoded.ChangedFiles[0].Path != "src/auth.js" {
		t.Errorf("expected src/auth.js, got %s", decoded.ChangedFiles[0].Path)
	}
	if len(decoded.ChangedPackages) != 1 || decoded.ChangedPackages[0] != "src/auth" {
		t.Errorf("expected packages [src/auth], got %v", decoded.ChangedPackages)
	}
}

func TestChangeSet_ShallowMetadata(t *testing.T) {
	t.Parallel()
	cs := &ChangeSet{
		Source:    "shallow-clone",
		IsShallow: true,
		Limitations: []string{"shallow clone: commit history may be incomplete"},
	}

	if !cs.IsShallow {
		t.Error("expected IsShallow to be true")
	}
	if len(cs.Limitations) != 1 {
		t.Errorf("expected 1 limitation, got %d", len(cs.Limitations))
	}
}
