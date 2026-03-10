package benchmark

import (
	"context"
	"fmt"
	"time"
)

// BenchResult holds all command results for a single repo.
type BenchResult struct {
	Repo     RepoMeta        `json:"repo"`
	Commands []CommandResult  `json:"commands"`
	RunAt    time.Time        `json:"runAt"`
}

// ProgressFunc is called after each command completes for streaming output.
type ProgressFunc func(cr CommandResult)

// RunBenchmark executes all commands against a single repo sequentially.
// Commands run sequentially within a repo to avoid resource contention
// (each hamlet invocation scans the full codebase). Repos can be
// parallelized at the caller level.
//
// If onProgress is non-nil, it is called after each command completes.
func RunBenchmark(ctx context.Context, hamletBin string, meta *RepoMeta, primary, debug []CommandSpec, onProgress ProgressFunc) BenchResult {
	result := BenchResult{
		Repo:  *meta,
		RunAt: time.Now(),
	}

	for _, spec := range primary {
		if spec.NeedGit && !meta.IsGitRepo {
			cr := CommandResult{
				Command:  spec.Name,
				RepoName: meta.Name,
				Error:    "skipped: not a git repo",
			}
			result.Commands = append(result.Commands, cr)
			if onProgress != nil {
				onProgress(cr)
			}
			continue
		}

		var cr CommandResult
		if spec.Name == "explain" {
			cr = RunExplain(ctx, hamletBin, meta.AbsPath)
		} else {
			cr = RunCommand(ctx, hamletBin, meta.AbsPath, spec)
		}
		cr.RepoName = meta.Name
		result.Commands = append(result.Commands, cr)
		if onProgress != nil {
			onProgress(cr)
		}
	}

	for _, spec := range debug {
		cr := RunCommand(ctx, hamletBin, meta.AbsPath, spec)
		cr.RepoName = meta.Name
		result.Commands = append(result.Commands, cr)
		if onProgress != nil {
			onProgress(cr)
		}
	}

	return result
}

// DefaultProgress prints each command result as it completes.
func DefaultProgress(cr CommandResult) {
	status := "OK"
	if cr.ExitCode != 0 || cr.Error != "" {
		status = fmt.Sprintf("FAIL (exit %d)", cr.ExitCode)
	}
	fmt.Printf("  %-20s %s  %dms\n", cr.Command, status, cr.RuntimeMs)
}
