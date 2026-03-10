package models

import "time"

// RepositoryMetadata represents top-level context about the repository
// being analyzed.
//
// This is intentionally product-facing, not just implementation-facing.
// The purpose of this model is to capture the basic identity and
// environment of the codebase so Hamlet can:
//   - describe what it analyzed
//   - serialize snapshots consistently
//   - support future historical comparisons
//   - provide context for benchmark and risk interpretation
type RepositoryMetadata struct {
	// Name is the repository or project name.
	Name string `json:"name"`

	// RootPath is the filesystem root that Hamlet analyzed.
	RootPath string `json:"rootPath"`

	// Languages lists the primary languages detected in the repository.
	Languages []string `json:"languages,omitempty"`

	// PackageManagers lists package/build managers observed in the repo.
	// Examples: npm, pnpm, yarn, pip, poetry, maven, gradle.
	PackageManagers []string `json:"packageManagers,omitempty"`

	// CISystems lists CI systems detected from repository configuration
	// or metadata. Examples: github-actions, circleci, buildkite.
	CISystems []string `json:"ciSystems,omitempty"`

	// SnapshotTimestamp is the repository-scoped timestamp for this snapshot.
	// It is kept in sync with TestSuiteSnapshot.GeneratedAt for backward
	// compatibility with earlier consumers.
	SnapshotTimestamp time.Time `json:"snapshotTimestamp"`

	// CommitSHA is the git commit associated with the snapshot if available.
	CommitSHA string `json:"commitSha,omitempty"`

	// Branch is the current branch if known.
	Branch string `json:"branch,omitempty"`
}
