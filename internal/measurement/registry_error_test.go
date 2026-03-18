package measurement

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestMeasurementDefaultRegistry_ReturnsNoError(t *testing.T) {
	t.Parallel()
	_, err := DefaultRegistry()
	if err != nil {
		t.Fatalf("DefaultRegistry should succeed: %v", err)
	}
}

func TestMeasurementDefaultRegistry_DefinitionCountStable(t *testing.T) {
	t.Parallel()
	reg, err := DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	// Should have all 18 measurement definitions.
	count := len(reg.All())
	if count < 18 {
		t.Errorf("expected at least 18 definitions, got %d", count)
	}
}

func TestMeasurementRegistry_DuplicateIDReturnsError(t *testing.T) {
	t.Parallel()
	r := NewRegistry()

	def := Definition{
		ID:          "test.duplicate",
		Description: "first",
		Dimension:   DimensionHealth,
		Compute:     func(snap *models.TestSuiteSnapshot) Result { return Result{} },
	}

	if err := r.Register(def); err != nil {
		t.Fatalf("first Register should succeed: %v", err)
	}

	err := r.Register(def)
	if err == nil {
		t.Fatal("duplicate ID should return error, not nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention 'duplicate', got: %v", err)
	}
}

func TestMeasurementRegistry_DuplicateDoesNotPanic(t *testing.T) {
	t.Parallel()
	r := NewRegistry()

	def := Definition{
		ID:        "test.safe",
		Dimension: DimensionHealth,
		Compute:   func(snap *models.TestSuiteSnapshot) Result { return Result{} },
	}

	_ = r.Register(def)
	err := r.Register(def)
	if err == nil {
		t.Fatal("expected error for duplicate, got nil")
	}
	// If we got here, no panic occurred.
}
