// Package airun implements the artifact model for AI validation runs,
// including content hashing for reproducibility and replay support.
package airun

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pmclSF/terrain/internal/models"
)

// Artifact is the complete output of a terrain ai run.
// It contains everything needed to replay and verify a run.
type Artifact struct {
	// Version is the artifact schema version.
	Version string `json:"version"`

	// CreatedAt is the UTC timestamp of the run.
	CreatedAt string `json:"createdAt"`

	// Mode is "impacted", "full", or "dry-run".
	Mode string `json:"mode"`

	// Framework is the detected eval framework.
	Framework string `json:"framework"`

	// Command is the execution command.
	Command string `json:"command,omitempty"`

	// Selected scenarios with selection reasoning.
	Selected []ScenarioEntry `json:"selected"`

	// Skipped scenarios with skip reasoning.
	Skipped []ScenarioEntry `json:"skipped,omitempty"`

	// Hashes captures content hashes for reproducibility verification.
	Hashes ContentHashes `json:"hashes"`

	// Signals detected during the run.
	Signals []SignalEntry `json:"signals,omitempty"`

	// Decision is the CI decision summary.
	Decision Decision `json:"decision"`

	// ExitCode is the process exit code (0 = pass, 1 = block).
	ExitCode int `json:"exitCode"`
}

// ScenarioEntry represents a scenario in the run artifact.
type ScenarioEntry struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Capability string   `json:"capability,omitempty"`
	Category   string   `json:"category,omitempty"`
	Reason     string   `json:"reason"`
	Surfaces   []string `json:"surfaces,omitempty"`
	Path       string   `json:"path,omitempty"`
}

// ContentHashes captures file content hashes for reproducibility.
type ContentHashes struct {
	// Prompts maps surface ID → SHA256 of the file content at run time.
	Prompts map[string]string `json:"prompts,omitempty"`

	// Contexts maps surface ID → SHA256.
	Contexts map[string]string `json:"contexts,omitempty"`

	// Datasets maps surface ID → SHA256.
	Datasets map[string]string `json:"datasets,omitempty"`

	// ToolDefs maps surface ID → SHA256.
	ToolDefs map[string]string `json:"toolDefinitions,omitempty"`

	// Retrievals maps surface ID → SHA256.
	Retrievals map[string]string `json:"retrievalSurfaces,omitempty"`

	// Config is the SHA256 of .terrain/terrain.yaml if present.
	Config string `json:"config,omitempty"`

	// Model is the model/provider reference if detectable.
	Model string `json:"model,omitempty"`
}

// SignalEntry is a signal in the run artifact.
type SignalEntry struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Scenario    string `json:"scenario,omitempty"`
	Explanation string `json:"explanation"`
}

// Decision is the CI decision summary.
type Decision struct {
	Action  string `json:"action"` // "pass", "warn", "block"
	Reason  string `json:"reason"`
	Signals int    `json:"signals"`
	Blocked int    `json:"blocked"`
}

// ReplayResult is the output of replaying an artifact.
type ReplayResult struct {
	// Match indicates whether the replay matches the original.
	Match bool `json:"match"`

	// Mismatches lists specific differences found.
	Mismatches []Mismatch `json:"mismatches,omitempty"`

	// CurrentHashes are the hashes computed during replay.
	CurrentHashes ContentHashes `json:"currentHashes"`

	// CurrentScenarios are the scenarios selected during replay.
	CurrentScenarios int `json:"currentScenarios"`

	// OriginalScenarios is the count from the artifact.
	OriginalScenarios int `json:"originalScenarios"`
}

// Mismatch describes a specific difference between original and replay.
type Mismatch struct {
	Kind     string `json:"kind"`     // "hash", "scenario", "config", "signal"
	Surface  string `json:"surface"`  // surface ID or file path
	Original string `json:"original"` // value at run time
	Current  string `json:"current"`  // value at replay time
	Detail   string `json:"detail"`   // human-readable explanation
}

// ComputeHashes computes content hashes for all AI surfaces in a snapshot.
func ComputeHashes(root string, surfaces []models.CodeSurface) ContentHashes {
	h := ContentHashes{
		Prompts:    map[string]string{},
		Contexts:   map[string]string{},
		Datasets:   map[string]string{},
		ToolDefs:   map[string]string{},
		Retrievals: map[string]string{},
	}

	for _, cs := range surfaces {
		hash := hashFile(root, cs.Path)
		if hash == "" {
			continue
		}
		switch cs.Kind {
		case models.SurfacePrompt:
			h.Prompts[cs.SurfaceID] = hash
		case models.SurfaceContext:
			h.Contexts[cs.SurfaceID] = hash
		case models.SurfaceDataset:
			h.Datasets[cs.SurfaceID] = hash
		case models.SurfaceToolDef:
			h.ToolDefs[cs.SurfaceID] = hash
		case models.SurfaceRetrieval:
			h.Retrievals[cs.SurfaceID] = hash
		}
	}

	// Hash config file.
	configPath := filepath.Join(root, ".terrain", "terrain.yaml")
	if hash := hashFileAbs(configPath); hash != "" {
		h.Config = hash
	}

	return h
}

// Replay loads an artifact and compares it to current repo state.
func Replay(artifactPath, root string, currentSurfaces []models.CodeSurface, currentScenarioCount int) (*ReplayResult, error) {
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("reading artifact: %w", err)
	}
	var art Artifact
	if err := json.Unmarshal(data, &art); err != nil {
		return nil, fmt.Errorf("parsing artifact: %w", err)
	}

	currentHashes := ComputeHashes(root, currentSurfaces)
	result := &ReplayResult{
		Match:             true,
		CurrentHashes:     currentHashes,
		CurrentScenarios:  currentScenarioCount,
		OriginalScenarios: len(art.Selected),
	}

	// Compare scenario counts.
	if currentScenarioCount != len(art.Selected) {
		result.Match = false
		result.Mismatches = append(result.Mismatches, Mismatch{
			Kind:     "scenario",
			Surface:  "scenario_count",
			Original: fmt.Sprintf("%d", len(art.Selected)),
			Current:  fmt.Sprintf("%d", currentScenarioCount),
			Detail:   fmt.Sprintf("scenario count changed: %d → %d", len(art.Selected), currentScenarioCount),
		})
	}

	// Compare hashes.
	compareHashMaps(art.Hashes.Prompts, currentHashes.Prompts, "prompt", result)
	compareHashMaps(art.Hashes.Contexts, currentHashes.Contexts, "context", result)
	compareHashMaps(art.Hashes.Datasets, currentHashes.Datasets, "dataset", result)
	compareHashMaps(art.Hashes.ToolDefs, currentHashes.ToolDefs, "tool_definition", result)
	compareHashMaps(art.Hashes.Retrievals, currentHashes.Retrievals, "retrieval", result)

	// Compare config hash.
	if art.Hashes.Config != "" && currentHashes.Config != "" && art.Hashes.Config != currentHashes.Config {
		result.Match = false
		result.Mismatches = append(result.Mismatches, Mismatch{
			Kind:     "config",
			Surface:  ".terrain/terrain.yaml",
			Original: art.Hashes.Config[:12],
			Current:  currentHashes.Config[:12],
			Detail:   "terrain config changed since original run",
		})
	}

	return result, nil
}

// SaveArtifact writes an artifact to the standard location.
func SaveArtifact(root string, art *Artifact) (string, error) {
	dir := filepath.Join(root, ".terrain", "artifacts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	art.Version = "1"
	art.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(art, "", "  ")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "ai-run-latest.json")
	return path, os.WriteFile(path, data, 0o644)
}

// --- Helpers ---

func hashFile(root, relPath string) string {
	return hashFileAbs(filepath.Join(root, relPath))
}

func hashFileAbs(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func compareHashMaps(original, current map[string]string, kind string, result *ReplayResult) {
	// Collect all keys.
	allKeys := map[string]bool{}
	for k := range original {
		allKeys[k] = true
	}
	for k := range current {
		allKeys[k] = true
	}

	sorted := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, key := range sorted {
		orig, hasOrig := original[key]
		curr, hasCurr := current[key]

		if hasOrig && !hasCurr {
			result.Match = false
			result.Mismatches = append(result.Mismatches, Mismatch{
				Kind:    "hash",
				Surface: key,
				Original: truncHash(orig),
				Current: "(removed)",
				Detail:  fmt.Sprintf("%s surface removed since original run", kind),
			})
		} else if !hasOrig && hasCurr {
			result.Match = false
			result.Mismatches = append(result.Mismatches, Mismatch{
				Kind:    "hash",
				Surface: key,
				Original: "(absent)",
				Current: truncHash(curr),
				Detail:  fmt.Sprintf("new %s surface added since original run", kind),
			})
		} else if orig != curr {
			result.Match = false
			result.Mismatches = append(result.Mismatches, Mismatch{
				Kind:    "hash",
				Surface: key,
				Original: truncHash(orig),
				Current: truncHash(curr),
				Detail:  fmt.Sprintf("%s content changed since original run", kind),
			})
		}
	}
}

func truncHash(h string) string {
	if len(h) > 12 {
		return h[:12]
	}
	return h
}

// ShortHash returns a truncated hash for display.
func ShortHash(h string) string {
	return truncHash(h)
}

// FlattenHashMap returns all hashes sorted for deterministic display.
func FlattenHashMap(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]string, len(keys))
	for i, k := range keys {
		result[i] = k + "=" + truncHash(m[k])
	}
	return result
}

// TotalHashCount returns the total number of hashed surfaces.
func (h ContentHashes) TotalHashCount() int {
	return len(h.Prompts) + len(h.Contexts) + len(h.Datasets) +
		len(h.ToolDefs) + len(h.Retrievals)
}

// SurfaceIDsFromHashes collects surface IDs into a string for display.
func SurfaceIDsFromHashes(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		// Extract just the name part from "surface:path:name".
		parts := strings.SplitN(k, ":", 3)
		if len(parts) >= 3 {
			keys = append(keys, parts[2])
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
