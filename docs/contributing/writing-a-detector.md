# Writing a Detector

This guide explains how to add a new signal detector to Terrain.

## 1. Implement the Detector Interface

Create a new file in the appropriate domain package (e.g., `internal/quality/`, `internal/migration/`):

```go
package quality

import "github.com/pmclSF/terrain/internal/models"

type MyNewDetector struct {
    // Add fields only if the detector needs runtime config (e.g., RepoRoot).
    // Most detectors should be stateless.
}

func (d *MyNewDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
    var signals []models.Signal

    for _, tf := range snap.TestFiles {
        // Analyze the test file...
        if problemFound {
            signals = append(signals, models.Signal{
                Type:            "myNewSignalType",
                Category:        models.CategoryQuality,
                Severity:        models.SeverityMedium,
                Confidence:      0.7,
                Location:        models.SignalLocation{File: tf.Path},
                Explanation:     "Clear explanation of what was found.",
                SuggestedAction: "What the user should do about it.",
                Metadata:        map[string]any{"key": "value"},
            })
        }
    }

    return signals
}
```

## 2. Register the Signal Type

Add the signal type constant to `internal/signals/signal_types.go`:

```go
const SignalMyNewType models.SignalType = "myNewSignalType"
```

Add a catalog entry to `internal/signals/registry.go`:

```go
SignalMyNewType: {
    Type:        SignalMyNewType,
    Category:    models.CategoryQuality,
    Title:       "My New Signal",
    Description: "Describes what this signal means.",
},
```

## 3. Register the Detector

Add the registration to `internal/engine/registry.go` in `DefaultRegistry()`:

```go
r.Register(signals.DetectorRegistration{
    Meta: signals.DetectorMeta{
        ID:           "quality.my-new-signal",
        Domain:       signals.DomainQuality,
        EvidenceType: signals.EvidenceStructuralPattern,
        Description:  "Detect the new quality issue.",
        SignalTypes:   []models.SignalType{signals.SignalMyNewType},
    },
    Detector: &quality.MyNewDetector{},
})
```

**Placement matters:** If your detector depends on signals from other detectors, register it after them and set `DependsOnSignals: true`.

## 4. Write Tests

Create a test file in the same package:

```go
func TestMyNewDetector_HappyPath(t *testing.T) {
    snap := &models.TestSuiteSnapshot{
        TestFiles: []models.TestFile{
            {Path: "src/__tests__/example.test.js", /* ... */},
        },
    }

    d := &MyNewDetector{}
    signals := d.Detect(snap)

    if len(signals) != 1 {
        t.Fatalf("got %d signals, want 1", len(signals))
    }
    if signals[0].Type != "myNewSignalType" {
        t.Errorf("type = %s, want myNewSignalType", signals[0].Type)
    }
}

func TestMyNewDetector_NoMatch(t *testing.T) {
    snap := &models.TestSuiteSnapshot{}
    d := &MyNewDetector{}
    signals := d.Detect(snap)

    if len(signals) != 0 {
        t.Errorf("got %d signals, want 0", len(signals))
    }
}
```

## 5. Update the Signal Catalog Test

Add your new signal type to the expected list in `internal/signals/registry_test.go`.

## 6. Update Documentation

Add the signal to `docs/signal-catalog.md`.

## Checklist

- [ ] Detector implements `signals.Detector` interface
- [ ] Signal type constant in `signal_types.go`
- [ ] Catalog entry in `registry.go` (signal registry)
- [ ] Registration in `engine/registry.go` (detector registry)
- [ ] Tests for happy path, no-match, and edge cases
- [ ] Signal catalog test updated
- [ ] `docs/signal-catalog.md` updated
- [ ] `go test ./internal/...` passes
