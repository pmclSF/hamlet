package ownership

import "testing"

func TestNormalizeOwnerID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"@team-auth", "team-auth"},
		{"team-auth", "team-auth"},
		{"@org/team-platform", "org/team-platform"},
		{"  @team-auth  ", "team-auth"},
		{"  @ team-auth  ", "team-auth"},
		{"", ""},
		{"@", ""},
	}
	for _, tt := range tests {
		got := NormalizeOwnerID(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeOwnerID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestOwnershipAssignment_IsUnowned(t *testing.T) {
	a := OwnershipAssignment{}
	if !a.IsUnowned() {
		t.Error("empty assignment should be unowned")
	}

	a.Owners = []Owner{{ID: "team-a"}}
	if a.IsUnowned() {
		t.Error("assignment with owners should not be unowned")
	}
}

func TestOwnershipAssignment_PrimaryOwner(t *testing.T) {
	a := OwnershipAssignment{}
	if a.PrimaryOwnerID() != "unknown" {
		t.Errorf("unowned primary = %q, want %q", a.PrimaryOwnerID(), "unknown")
	}

	a.Owners = []Owner{{ID: "team-a"}, {ID: "team-b"}}
	if a.PrimaryOwnerID() != "team-a" {
		t.Errorf("primary = %q, want %q", a.PrimaryOwnerID(), "team-a")
	}
}

func TestOwnershipAssignment_HasOwner(t *testing.T) {
	a := OwnershipAssignment{
		Owners: []Owner{{ID: "team-a"}, {ID: "team-b"}},
	}
	if !a.HasOwner("team-a") {
		t.Error("should have team-a")
	}
	if !a.HasOwner("team-b") {
		t.Error("should have team-b")
	}
	if a.HasOwner("team-c") {
		t.Error("should not have team-c")
	}
}

func TestSourceConfidence(t *testing.T) {
	tests := []struct {
		source SourceType
		want   Confidence
	}{
		{SourceExplicitConfig, ConfidenceHigh},
		{SourceCodeowners, ConfidenceHigh},
		{SourcePackageMetadata, ConfidenceMedium},
		{SourcePathMapping, ConfidenceMedium},
		{SourceGitHistory, ConfidenceLow},
		{SourceDirectoryFallback, ConfidenceLow},
		{SourceUnknown, ConfidenceNone},
	}
	for _, tt := range tests {
		got := SourceConfidence(tt.source)
		if got != tt.want {
			t.Errorf("SourceConfidence(%q) = %q, want %q", tt.source, got, tt.want)
		}
	}
}
