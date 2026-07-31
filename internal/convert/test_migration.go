package convert

import (
	"fmt"
	"strings"
)

// ConversionInputError reports invalid user-supplied conversion parameters.
type ConversionInputError struct {
	message string
}

func (e ConversionInputError) Error() string {
	return e.message
}

// TestMigrationOptions controls one-off native test migration execution.
type TestMigrationOptions struct {
	Alias             string `json:"alias,omitempty"`
	From              string `json:"from,omitempty"`
	To                string `json:"to,omitempty"`
	Output            string `json:"output,omitempty"`
	PreserveStructure bool   `json:"preserveStructure,omitempty"`
	BatchSize         int    `json:"batchSize,omitempty"`
	Concurrency       int    `json:"concurrency,omitempty"`
	AutoDetect        bool   `json:"autoDetect,omitempty"`
	ValidateSyntax    bool   `json:"validateSyntax,omitempty"`
	ValidationMode    string `json:"validationMode,omitempty"`
	Plan              bool   `json:"plan,omitempty"`
	DryRun            bool   `json:"dryRun,omitempty"`

	// HistoryRoot, when set, points at the repository root that owns
	// `.terrain/conversion-history/`. The runtime appends one record
	// per successful conversion. Empty disables history (preserves
	// pre-0.2 behavior for callers that haven't opted in).
	HistoryRoot string `json:"historyRoot,omitempty"`

	// TerrainVersion is stamped into the history record so audit
	// readers know which engine produced the conversion. Plumbed
	// from the CLI's main.version build var.
	TerrainVersion string `json:"terrainVersion,omitempty"`

	// Preview runs the conversion to a temp directory and returns
	// per-file unified diffs without writing to the user's --output.
	// Distinct from DryRun (which produces a structured plan only):
	// Preview shows the actual converted content as a diff against
	// the source, useful when the structural plan is fine but you
	// want to eyeball the output before committing.
	Preview bool `json:"preview,omitempty"`
}

// TestMigrationPlan describes a native conversion plan or dry-run preview.
type TestMigrationPlan struct {
	Command         string     `json:"command"`
	Mode            string     `json:"mode"`
	Source          string     `json:"source"`
	Output          string     `json:"output,omitempty"`
	Alias           string     `json:"alias,omitempty"`
	ValidationMode  string     `json:"validationMode,omitempty"`
	Direction       Direction  `json:"direction"`
	SourceDetection *Detection `json:"sourceDetection,omitempty"`
	ExecutionStatus string     `json:"executionStatus"`
	NextStep        string     `json:"nextStep"`
}

// TestMigrationResult is the native Terrain result for a single test migration request.
type TestMigrationResult struct {
	Source          string             `json:"source"`
	Alias           string             `json:"alias,omitempty"`
	ValidationMode  string             `json:"validationMode,omitempty"`
	Direction       Direction          `json:"direction"`
	SourceDetection *Detection         `json:"sourceDetection,omitempty"`
	Plan            *TestMigrationPlan `json:"plan,omitempty"`
	Execution       *ExecutionResult   `json:"execution,omitempty"`

	// Preview is populated when TestMigrationOptions.Preview was set.
	// One entry per converted file with the unified diff against the
	// original source. Mutually exclusive with Execution: preview runs
	// to a temp directory and the temp output is discarded after the
	// diff is captured.
	Preview []FilePreview `json:"preview,omitempty"`
}

// FilePreview is one converted file's diff captured during a Preview run.
type FilePreview struct {
	SourcePath string `json:"sourcePath"`
	OutputPath string `json:"outputPath,omitempty"`
	Status     string `json:"status,omitempty"`
	Changed    bool   `json:"changed"`
	Diff       string `json:"diff"`
}

// RunTestMigration plans or executes a single native test migration request.
func RunTestMigration(source string, options TestMigrationOptions) (TestMigrationResult, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return TestMigrationResult{}, inputErrorf("source is required")
	}

	direction, detection, alias, err := resolveTestMigrationDirection(source, options)
	if err != nil {
		return TestMigrationResult{}, err
	}

	result := TestMigrationResult{
		Source:          source,
		Alias:           alias,
		ValidationMode:  string(normalizeValidationMode(options.ValidationMode)),
		Direction:       direction,
		SourceDetection: detection,
	}

	if options.Plan || options.DryRun {
		plan := buildTestMigrationPlan(source, direction, detection, alias, options)
		result.Plan = &plan
		return result, nil
	}

	if direction.GoNativeState != GoNativeStateImplemented &&
		direction.GoNativeState != GoNativeStateExperimental {
		return result, fmt.Errorf(
			"go-native conversion execution for %s -> %s is not implemented yet; use plan mode to inspect the migration path",
			direction.From,
			direction.To,
		)
	}

	// Preview mode: run the conversion to a temp directory, build
	// per-file unified diffs, then discard the temp output. The user's
	// --output (if set) is ignored — preview is read-only.
	if options.Preview {
		previews, err := runPreview(source, direction, options)
		if err != nil {
			return result, err
		}
		result.Preview = previews
		return result, nil
	}

	execution, err := Execute(source, direction, ExecuteOptions{
		Output:            options.Output,
		PreserveStructure: options.PreserveStructure,
		BatchSize:         options.BatchSize,
		Concurrency:       options.Concurrency,
	})
	if err != nil {
		return result, err
	}
	// 0.2 per-file confidence: walk Files, compute heuristic
	// covered/lossy/confidence for each (source, output) pair. The
	// metrics surface in JSON output and feed the report renderer.
	annotateFileConfidence(&execution)

	validationMode := normalizeValidationMode(options.ValidationMode)
	execution.ValidationMode = string(validationMode)
	validationErr := ValidateExecutionResultForDirection(execution, direction)
	switch validationMode {
	case ValidationModeStrict:
		if validationErr != nil {
			if cleanupErr := CleanupExecutionOutputs(execution); cleanupErr != nil {
				return result, fmt.Errorf("%v (cleanup failed: %w)", validationErr, cleanupErr)
			}
			return result, validationErr
		}
		execution.Validated = true
	case ValidationModeBestEffort:
		execution.Warnings = append(execution.Warnings, validationWarningsForError(validationMode, validationErr)...)
		execution.Validated = validationErr == nil
	}

	// 0.2 conversion history: append AFTER validation so the record
	// reflects the final Validated state. Errors here do not fail
	// the conversion — by the time we reach this point the user's
	// output is already on disk and an audit-log failure shouldn't
	// undo their successful run.
	if options.HistoryRoot != "" {
		rec := HistoryRecordFromExecution(execution, options.TerrainVersion)
		if err := AppendConversionHistory(options.HistoryRoot, rec); err != nil {
			execution.Warnings = append(execution.Warnings,
				fmt.Sprintf("conversion history append failed: %v", err))
		}
	}

	result.Execution = &execution
	return result, nil
}

func resolveTestMigrationDirection(source string, options TestMigrationOptions) (Direction, *Detection, string, error) {
	from := NormalizeFramework(options.From)
	to := NormalizeFramework(options.To)
	alias := strings.ToLower(strings.TrimSpace(options.Alias))

	if alias != "" {
		aliasDirection, ok := LookupShorthand(alias)
		if !ok {
			return Direction{}, nil, alias, inputErrorf("unknown shorthand: %s", alias)
		}
		if from != "" && from != aliasDirection.From {
			return Direction{}, nil, alias, inputErrorf(
				"shorthand %s maps to %s -> %s, not %s -> %s",
				alias,
				aliasDirection.From,
				aliasDirection.To,
				from,
				emptyOrUnknown(to),
			)
		}
		if to != "" && to != aliasDirection.To {
			return Direction{}, nil, alias, inputErrorf(
				"shorthand %s maps to %s -> %s, not %s -> %s",
				alias,
				aliasDirection.From,
				aliasDirection.To,
				emptyOrUnknown(from),
				to,
			)
		}
		if from == "" {
			from = aliasDirection.From
		}
		if to == "" {
			to = aliasDirection.To
		}
	}

	var detection *Detection
	if from == "" && options.AutoDetect {
		detected, err := DetectSource(source)
		if err != nil {
			return Direction{}, nil, alias, err
		}
		detection = &detected
		if detected.Framework == "" || detected.Framework == "unknown" {
			return Direction{}, nil, alias, fmt.Errorf("could not auto-detect source framework from %s", source)
		}
		if detected.Mode == "directory" && !detected.AutoDetectSafe {
			return Direction{}, detection, alias, fmt.Errorf(
				"auto-detect is not safe for %s (%s); use --from explicitly. Candidates: %s",
				source,
				detected.Recommendation,
				formatDetectionCandidates(detected.Candidates),
			)
		}
		from = detected.Framework
	}

	if from == "" {
		return Direction{}, detection, alias, inputErrorf("source framework is required unless auto-detect is enabled")
	}
	if to == "" {
		return Direction{}, detection, alias, inputErrorf("target framework is required")
	}
	direction, err := resolveSupportedDirection(from, to)
	if err != nil {
		return Direction{}, detection, alias, err
	}
	return direction, detection, alias, nil
}

func buildTestMigrationPlan(source string, direction Direction, detection *Detection, alias string, options TestMigrationOptions) TestMigrationPlan {
	mode := "plan"
	if options.DryRun && !options.Plan {
		mode = "dry-run"
	}

	plan := TestMigrationPlan{
		Command:         "convert",
		Mode:            mode,
		Source:          source,
		Output:          options.Output,
		Alias:           alias,
		ValidationMode:  string(normalizeValidationMode(options.ValidationMode)),
		Direction:       direction,
		SourceDetection: detection,
		ExecutionStatus: "cataloged-not-executable",
		NextStep:        "The Go CLI owns the conversion catalog, shorthands, and detection contract. This direction is cataloged but not executable.",
	}
	switch direction.GoNativeState {
	case GoNativeStateImplemented:
		plan.ExecutionStatus = "executable"
		plan.NextStep = "Run the same request without plan mode to execute the Go-native converter for this direction."
	case GoNativeStateExperimental:
		plan.ExecutionStatus = "executable-experimental"
		plan.NextStep = "Run the same request without plan mode to execute the experimental Go-native converter; expect manual cleanup post-conversion."
	}
	return plan
}

func inputErrorf(format string, args ...any) error {
	return ConversionInputError{message: fmt.Sprintf(format, args...)}
}

func resolveSupportedDirection(from, to string) (Direction, error) {
	from = NormalizeFramework(from)
	to = NormalizeFramework(to)

	if _, ok := LookupFramework(from); !ok {
		return Direction{}, inputErrorf(
			"invalid source framework: %s. Valid options: %s",
			from,
			strings.Join(FrameworkNames(), ", "),
		)
	}
	if _, ok := LookupFramework(to); !ok {
		return Direction{}, inputErrorf(
			"invalid target framework: %s. Valid options: %s",
			to,
			strings.Join(FrameworkNames(), ", "),
		)
	}
	if from == to {
		return Direction{}, inputErrorf("source and target frameworks must be different")
	}

	direction, ok := LookupDirection(from, to)
	if !ok {
		targets := SupportedTargets(from)
		if len(targets) == 0 {
			return Direction{}, inputErrorf("unsupported source framework: %s", from)
		}
		return Direction{}, inputErrorf(
			"unsupported conversion: %s to %s. Supported targets for %s: %s",
			from,
			to,
			from,
			strings.Join(targets, ", "),
		)
	}
	return direction, nil
}

func emptyOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func formatDetectionCandidates(candidates []DetectionCandidate) string {
	if len(candidates) == 0 {
		return "none"
	}

	parts := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		parts = append(parts, fmt.Sprintf("%s (%.0f%%, %d file(s), %.0f%% share)", candidate.Framework, candidate.Confidence*100, candidate.Files, candidate.FileShare*100))
	}
	return strings.Join(parts, ", ")
}
