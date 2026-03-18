// Command terrain-bench runs Terrain CLI commands against local repositories,
// assesses output quality, and generates benchmark reports.
//
// Usage:
//
//	terrain-bench
//	terrain-bench --repo terrain --command analyze
//	terrain-bench --discover benchmarks/repos/
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pmclSF/terrain/internal/benchmark"
)

// repoResult bundles a benchmark result with its assessment for collection.
type repoResult struct {
	index      int
	result     benchmark.BenchResult
	assessment benchmark.RepoAssessment
}

func main() {
	var (
		configPath         string
		repoFilter         string
		cmdFilter          string
		discoverDir        string
		terrainBin         string
		outputDir          string
		timeout            int
		sequential         bool
		rebuildBinary      bool
		heavyTimeoutFactor float64
		stdoutCaptureKB    int
		stderrCaptureKB    int
	)

	flag.StringVar(&configPath, "config", "benchmarks/repos.json", "path to benchmark config file")
	flag.StringVar(&repoFilter, "repo", "", "run only this repo (by name)")
	flag.StringVar(&cmdFilter, "command", "", "run only this command (e.g., analyze, impact)")
	flag.StringVar(&discoverDir, "discover", "", "auto-discover repos from this directory")
	flag.StringVar(&terrainBin, "terrain", "", "path to terrain binary (default: auto-detect)")
	flag.StringVar(&outputDir, "output", "benchmarks/output", "output directory for results")
	flag.IntVar(&timeout, "timeout", 120, "per-command timeout in seconds")
	flag.BoolVar(&sequential, "sequential", false, "run repos sequentially instead of in parallel")
	flag.BoolVar(&rebuildBinary, "rebuild", true, "rebuild terrain binary from source before running (ignored when --terrain is set)")
	flag.Float64Var(&heavyTimeoutFactor, "heavy-timeout-factor", 4.0, "timeout multiplier for heavy commands (insights/debug)")
	flag.IntVar(&stdoutCaptureKB, "stdout-capture-kb", 512, "max stdout to retain per command (KB)")
	flag.IntVar(&stderrCaptureKB, "stderr-capture-kb", 128, "max stderr to retain per command (KB)")
	flag.Parse()

	// Find terrain binary.
	if terrainBin == "" {
		terrainBin = findTerrainBinary(rebuildBinary)
	}
	if terrainBin == "" {
		fmt.Fprintf(os.Stderr, "error: terrain binary not found. Build with: go build -o /tmp/terrain-bench ./cmd/terrain/\n")
		os.Exit(1)
	}

	// Find project root.
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Load repos.
	var repos []benchmark.Repo
	if discoverDir != "" {
		repos, err = benchmark.DiscoverRepos(discoverDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error discovering repos: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfgPath := configPath
		if !filepath.IsAbs(cfgPath) {
			cfgPath = filepath.Join(projectRoot, cfgPath)
		}
		repos, err = benchmark.LoadBenchmarkRepos(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			os.Exit(1)
		}
	}

	// Filter repos.
	if repoFilter != "" {
		var filtered []benchmark.Repo
		for _, r := range repos {
			if r.Name == repoFilter {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "error: repo %q not found in config\n", repoFilter)
			os.Exit(1)
		}
		repos = filtered
	}

	// Detect available commands.
	primary, debug, err := benchmark.DetectCommands(terrainBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error detecting commands: %v\n", err)
		os.Exit(1)
	}

	// Filter commands.
	if cmdFilter != "" {
		primary = filterSpecs(primary, cmdFilter)
		debug = filterSpecs(debug, cmdFilter)
	}

	// Set per-command timeout and capture limits.
	benchmark.PerCommandTimeout = time.Duration(timeout) * time.Second
	if heavyTimeoutFactor < 1.0 {
		heavyTimeoutFactor = 1.0
	}
	benchmark.HeavyCommandTimeoutFactor = heavyTimeoutFactor
	if stdoutCaptureKB < 64 {
		stdoutCaptureKB = 64
	}
	if stderrCaptureKB < 16 {
		stderrCaptureKB = 16
	}
	benchmark.MaxStdoutCaptureBytes = stdoutCaptureKB * 1024
	benchmark.MaxStderrCaptureBytes = stderrCaptureKB * 1024

	mode := "parallel"
	if sequential {
		mode = "sequential"
	}

	fmt.Printf("Terrain CLI Benchmark\n")
	fmt.Printf("====================\n")
	fmt.Printf("Binary:   %s\n", terrainBin)
	fmt.Printf("Repos:    %d (%s)\n", len(repos), mode)
	fmt.Printf("Commands: %d primary + %d debug (sequential per repo)\n", len(primary), len(debug))
	fmt.Printf("Timeout:  %ds per command\n\n", timeout)

	ctx := context.Background()

	// Resolve all repos upfront.
	type resolvedRepo struct {
		index int
		meta  *benchmark.RepoMeta
	}
	var resolved []resolvedRepo
	for i, repoCfg := range repos {
		meta, err := benchmark.ResolveRepo(repoCfg, projectRoot)
		if err != nil {
			fmt.Printf("SKIP %s: %v\n", repoCfg.Name, err)
			continue
		}
		resolved = append(resolved, resolvedRepo{index: i, meta: meta})
	}

	// Collect results (ordered by original index).
	allRepoResults := make([]repoResult, len(resolved))

	if sequential || len(resolved) == 1 {
		// Sequential execution.
		for i, rr := range resolved {
			fmt.Printf("--- %s (%s) ---\n", rr.meta.Name, rr.meta.AbsPath)
			br := benchmark.RunBenchmark(ctx, terrainBin, rr.meta, primary, debug, benchmark.DefaultProgress)
			assessment := benchmark.AssessResults(br)
			fmt.Printf("  Overall credibility: %d\n\n", assessment.OverallScore)
			allRepoResults[i] = repoResult{index: rr.index, result: br, assessment: assessment}
		}
	} else {
		// Parallel execution across repos.
		var mu sync.Mutex
		var wg sync.WaitGroup
		for i, rr := range resolved {
			wg.Add(1)
			go func(idx int, meta *benchmark.RepoMeta) {
				defer wg.Done()
				progress := func(cr benchmark.CommandResult) {
					mu.Lock()
					defer mu.Unlock()
					benchmark.DefaultProgress(cr)
				}
				br := benchmark.RunBenchmark(ctx, terrainBin, meta, primary, debug, progress)
				assessment := benchmark.AssessResults(br)

				mu.Lock()
				fmt.Printf("  [%s] Overall credibility: %d\n\n", meta.Name, assessment.OverallScore)
				allRepoResults[idx] = repoResult{index: idx, result: br, assessment: assessment}
				mu.Unlock()
			}(i, rr.meta)
		}
		wg.Wait()
	}

	// Collect ordered results.
	var allResults []benchmark.BenchResult
	var allAssessments []benchmark.RepoAssessment
	for _, rr := range allRepoResults {
		allResults = append(allResults, rr.result)
		allAssessments = append(allAssessments, rr.assessment)
	}

	// Write output.
	outDir := outputDir
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(projectRoot, outDir)
	}

	if err := benchmark.WriteResults(outDir, allResults, allAssessments); err != nil {
		fmt.Fprintf(os.Stderr, "error writing results: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Results written to %s/\n", outputDir)
	fmt.Printf("  benchmark-results.json\n")
	fmt.Printf("  benchmark-report.md\n")
	fmt.Printf("  cli-benchmark-assessment.json\n")
}

func findTerrainBinary(rebuild bool) string {
	if rebuild {
		fmt.Println("Building terrain binary...")
		cmd := exec.Command("go", "build", "-o", "/tmp/terrain-bench", "./cmd/terrain/")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return "/tmp/terrain-bench"
		}
		fmt.Fprintln(os.Stderr, "warning: failed to rebuild; falling back to existing binary lookup")
	}

	candidates := []string{
		"/tmp/terrain-bench",
		"./terrain",
	}

	if p, err := exec.LookPath("terrain"); err == nil {
		candidates = append([]string{p}, candidates...)
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return ""
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

func filterSpecs(specs []benchmark.CommandSpec, filter string) []benchmark.CommandSpec {
	var filtered []benchmark.CommandSpec
	for _, s := range specs {
		if s.Name == filter || strings.HasPrefix(s.Name, filter+":") {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
