package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/logging"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/migration"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/reporting"
)

func runMigration(subCmd, root string, jsonOutput bool, file, scope string) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	switch subCmd {
	case "readiness":
		readiness := migration.ComputeReadiness(result.Snapshot)
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(readiness)
		}
		reporting.RenderMigrationReport(os.Stdout, readiness)
		return nil

	case "blockers":
		readiness := migration.ComputeReadiness(result.Snapshot)
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{
				"totalBlockers":          readiness.TotalBlockers,
				"blockersByType":         readiness.BlockersByType,
				"representativeBlockers": readiness.RepresentativeBlockers,
				"areaAssessments":        readiness.AreaAssessments,
			})
		}
		reporting.RenderMigrationBlockers(os.Stdout, readiness)
		return nil

	case "preview":
		if file != "" {
			preview := migration.PreviewFile(result.Snapshot, file, absRoot)
			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(preview)
			}
			reporting.RenderMigrationPreview(os.Stdout, preview)
			return nil
		}
		// Scope-based preview
		previews := migration.PreviewScope(result.Snapshot, scope, absRoot)
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(previews)
		}
		reporting.RenderMigrationPreviewScope(os.Stdout, previews)
		return nil

	default:
		return fmt.Errorf("unknown migration subcommand: %q (valid: readiness, blockers, preview)", subCmd)
	}
}

// runExportBenchmark performs analysis and outputs a benchmark-safe JSON export.
func runExportBenchmark(root string) error {
	result, err := engine.RunPipeline(root, defaultPipelineOptions())
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	ms := metrics.Derive(result.Snapshot)
	export := benchmark.BuildExport(result.Snapshot, ms, result.HasPolicy)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(export)
}

// runCompare loads two snapshots and produces a comparison report.
//
// If --from and --to are not specified, it looks for the two most recent
// snapshots in .terrain/snapshots/.
func runCompare(fromPath, toPath, root string, jsonOutput bool) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	// Resolve snapshot paths if not explicitly provided.
	if fromPath == "" || toPath == "" {
		snapDir := filepath.Join(absRoot, ".terrain", "snapshots")
		latest, previous, err := findRecentSnapshots(snapDir)
		if err != nil {
			return err
		}
		if toPath == "" {
			toPath = latest
		}
		if fromPath == "" {
			fromPath = previous
		}
	}

	if fromPath == "" || toPath == "" {
		return fmt.Errorf("need at least two snapshots to compare; use --write-snapshot with terrain analyze first")
	}

	fromSnap, err := loadSnapshot(fromPath)
	if err != nil {
		return fmt.Errorf("failed to load baseline snapshot: %w", err)
	}
	toSnap, err := loadSnapshot(toPath)
	if err != nil {
		return fmt.Errorf("failed to load current snapshot: %w", err)
	}

	comp := comparison.Compare(fromSnap, toSnap)

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(comp)
	}

	reporting.RenderComparisonReport(os.Stdout, comp)
	return nil
}

func loadSnapshot(path string) (*models.TestSuiteSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap models.TestSuiteSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("invalid snapshot JSON in %s: %w", path, err)
	}
	models.MigrateSnapshotInPlace(&snap)
	return &snap, nil
}

// findRecentSnapshots returns the two most recent snapshot files in the directory.
func findRecentSnapshots(dir string) (latest, previous string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", fmt.Errorf("no snapshot history found. Run `terrain analyze --write-snapshot` to begin tracking")
	}

	var snapFiles []string
	for _, e := range entries {
		name := e.Name()
		if name == "latest.json" || !strings.HasSuffix(name, ".json") {
			continue
		}
		snapFiles = append(snapFiles, filepath.Join(dir, name))
	}

	sort.Strings(snapFiles) // Timestamped names sort chronologically

	if len(snapFiles) < 2 {
		latestPath := filepath.Join(dir, "latest.json")
		if _, statErr := os.Stat(latestPath); statErr == nil && len(snapFiles) == 1 {
			return latestPath, snapFiles[0], nil
		}
		return "", "", fmt.Errorf("need at least 2 snapshots to compare; found %d. Run `terrain analyze --write-snapshot` to save snapshots", len(snapFiles))
	}

	return snapFiles[len(snapFiles)-1], snapFiles[len(snapFiles)-2], nil
}

// persistSnapshot writes the snapshot to .terrain/snapshots/ as both
// latest.json and a timestamped archive file.
func persistSnapshot(snapshot *models.TestSuiteSnapshot, root string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	dir := filepath.Join(absRoot, ".terrain", "snapshots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	latestPath := filepath.Join(dir, "latest.json")
	if err := os.WriteFile(latestPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}

	ts := snapshot.GeneratedAt.UTC().Format("2006-01-02T15-04-05Z")
	archivePath := filepath.Join(dir, ts+".json")
	if err := os.WriteFile(archivePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write archive snapshot: %w", err)
	}

	logging.L().Info("snapshot persisted", "latest", latestPath, "archive", archivePath)
	return nil
}

