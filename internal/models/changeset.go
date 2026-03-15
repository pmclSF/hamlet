package models

import (
	"encoding/json"
	"time"
)

// ChangeSet is a normalized, reusable representation of a code change.
//
// It captures what changed between two points in a repository: files, packages,
// and optionally services and config artifacts. ChangeSet is the canonical
// input to impact analysis, PR analysis, and any other reasoning that starts
// from "what changed?"
//
// Design principle: ChangeSet is a value object — it describes state, not
// behavior. Construction is handled by dedicated builder functions
// (e.g., ChangeSetFromGitDiff). Consumers read the fields directly.
type ChangeSet struct {
	// Repository is the repository identifier (path or URL).
	Repository string `json:"repository,omitempty"`

	// BaseSHA is the base commit SHA (the "before" state).
	// Empty when history is unavailable (shallow clone, working tree diff).
	BaseSHA string `json:"baseSha,omitempty"`

	// HeadSHA is the head commit SHA (the "after" state).
	// Empty when diffing against the working tree.
	HeadSHA string `json:"headSha,omitempty"`

	// BaseRef is the symbolic ref used as the baseline (e.g., "main", "HEAD~1").
	// Preserved for display and debugging; BaseSHA is authoritative.
	BaseRef string `json:"baseRef,omitempty"`

	// Source describes how the ChangeSet was constructed.
	// Values: "git-diff", "git-diff-working-tree", "explicit", "ci-changed-files",
	// "snapshot-compare", "shallow-clone"
	Source string `json:"source"`

	// ChangedFiles lists all changed files with their change kind.
	ChangedFiles []ChangedFile `json:"changedFiles"`

	// ChangedPackages lists packages (Go packages, npm packages, Python modules)
	// that contain changed files. Derived from file paths using language-aware
	// package resolution.
	ChangedPackages []string `json:"changedPackages,omitempty"`

	// ChangedServices lists service names inferred from changed file paths,
	// when the repository follows a recognizable service layout (e.g.,
	// services/<name>/, cmd/<name>/, apps/<name>/).
	ChangedServices []string `json:"changedServices,omitempty"`

	// ChangedConfigs lists changed configuration or generated artifact paths
	// (CI configs, Dockerfiles, Makefiles, terraform, etc.).
	ChangedConfigs []string `json:"changedConfigs,omitempty"`

	// IsShallow indicates that git history was limited (shallow clone or
	// missing base ref). When true, some fields (BaseSHA, ChangedPackages)
	// may be incomplete or approximated.
	IsShallow bool `json:"isShallow,omitempty"`

	// Limitations describes data gaps affecting the ChangeSet.
	// Examples: "shallow clone: base SHA unavailable",
	// "base ref not found: falling back to working tree diff"
	Limitations []string `json:"limitations,omitempty"`

	// CreatedAt is when this ChangeSet was constructed.
	CreatedAt time.Time `json:"createdAt"`
}

// ChangedFile represents a single changed file within a ChangeSet.
type ChangedFile struct {
	Path       string     `json:"path"`
	ChangeKind ChangeKind `json:"changeKind"`

	// OldPath is set when ChangeKind is "renamed".
	OldPath string `json:"oldPath,omitempty"`

	// IsTestFile indicates if this file is a test file.
	IsTestFile bool `json:"isTestFile"`
}

// ChangeKind describes how a file was changed.
type ChangeKind string

const (
	ChangeAdded    ChangeKind = "added"
	ChangeModified ChangeKind = "modified"
	ChangeDeleted  ChangeKind = "deleted"
	ChangeRenamed  ChangeKind = "renamed"
)

// SourceFiles returns only non-test changed files.
func (cs *ChangeSet) SourceFiles() []ChangedFile {
	var out []ChangedFile
	for _, f := range cs.ChangedFiles {
		if !f.IsTestFile {
			out = append(out, f)
		}
	}
	return out
}

// TestFiles returns only test changed files.
func (cs *ChangeSet) TestFiles() []ChangedFile {
	var out []ChangedFile
	for _, f := range cs.ChangedFiles {
		if f.IsTestFile {
			out = append(out, f)
		}
	}
	return out
}

// FileCount returns the total number of changed files.
func (cs *ChangeSet) FileCount() int {
	return len(cs.ChangedFiles)
}

// HasFile returns true if the given path is in the ChangeSet.
func (cs *ChangeSet) HasFile(path string) bool {
	for _, f := range cs.ChangedFiles {
		if f.Path == path {
			return true
		}
	}
	return false
}

// MarshalJSON implements json.Marshaler with deterministic output.
func (cs *ChangeSet) MarshalJSON() ([]byte, error) {
	type Alias ChangeSet
	return json.Marshal((*Alias)(cs))
}
