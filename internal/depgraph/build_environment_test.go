package depgraph

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestBuildEnvironments_CreatesNodes(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Environments: []models.Environment{
			{
				EnvironmentID: "env:ci-linux-node22",
				Name:          "CI Linux Node 22",
				OS:            "linux",
				OSVersion:     "22.04",
				Runtime:       "node-22",
				CIProvider:    "github-actions",
				InferredFrom:  "github-actions:matrix",
			},
			{
				EnvironmentID: "env:ci-macos-node22",
				Name:          "CI macOS Node 22",
				OS:            "macos",
				Runtime:       "node-22",
				CIProvider:    "github-actions",
			},
		},
	}

	g := Build(snap)

	if g.Node("env:ci-linux-node22") == nil {
		t.Fatal("expected env:ci-linux-node22 node")
	}
	if g.Node("env:ci-macos-node22") == nil {
		t.Fatal("expected env:ci-macos-node22 node")
	}

	n := g.Node("env:ci-linux-node22")
	if n.Type != NodeEnvironment {
		t.Errorf("expected environment type, got %s", n.Type)
	}
	if n.Name != "CI Linux Node 22" {
		t.Errorf("expected name 'CI Linux Node 22', got %q", n.Name)
	}
	if n.Metadata["os"] != "linux" {
		t.Errorf("expected os=linux, got %q", n.Metadata["os"])
	}
	if n.Metadata["runtime"] != "node-22" {
		t.Errorf("expected runtime=node-22, got %q", n.Metadata["runtime"])
	}
	if n.Metadata["inferredFrom"] != "github-actions:matrix" {
		t.Errorf("expected inferredFrom metadata, got %q", n.Metadata["inferredFrom"])
	}
}

func TestBuildEnvironments_SkipsEmptyID(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Environments: []models.Environment{
			{EnvironmentID: "", Name: "missing ID"},
			{EnvironmentID: "env:valid", Name: "valid"},
		},
	}

	g := Build(snap)

	envNodes := g.NodesByType(NodeEnvironment)
	if len(envNodes) != 1 {
		t.Errorf("expected 1 environment node, got %d", len(envNodes))
	}
}

func TestBuildEnvironments_ProductionMetadata(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Environments: []models.Environment{
			{
				EnvironmentID: "env:staging",
				Name:          "Staging",
				IsProduction:  true,
				ResourceClass: "xlarge",
			},
		},
	}

	g := Build(snap)

	n := g.Node("env:staging")
	if n == nil {
		t.Fatal("expected env:staging node")
	}
	if n.Metadata["isProduction"] != "true" {
		t.Error("expected isProduction=true in metadata")
	}
	if n.Metadata["resourceClass"] != "xlarge" {
		t.Errorf("expected resourceClass=xlarge, got %q", n.Metadata["resourceClass"])
	}
}

func TestBuildEnvironmentClasses_CreatesNodesAndEdges(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Environments: []models.Environment{
			{EnvironmentID: "env:ci-linux", Name: "Linux"},
			{EnvironmentID: "env:ci-macos", Name: "macOS"},
			{EnvironmentID: "env:ci-windows", Name: "Windows"},
		},
		EnvironmentClasses: []models.EnvironmentClass{
			{
				ClassID:   "envclass:os",
				Name:      "Operating Systems",
				Dimension: "os",
				MemberIDs: []string{"env:ci-linux", "env:ci-macos", "env:ci-windows"},
			},
		},
	}

	g := Build(snap)

	classNode := g.Node("envclass:os")
	if classNode == nil {
		t.Fatal("expected envclass:os node")
	}
	if classNode.Type != NodeEnvironmentClass {
		t.Errorf("expected environment_class type, got %s", classNode.Type)
	}
	if classNode.Metadata["dimension"] != "os" {
		t.Errorf("expected dimension=os, got %q", classNode.Metadata["dimension"])
	}

	// Check edges from class to members.
	outgoing := g.Outgoing("envclass:os")
	containsEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeEnvironmentClassContains {
			containsEdges++
		}
	}
	if containsEdges != 3 {
		t.Errorf("expected 3 contains edges, got %d", containsEdges)
	}
}

func TestBuildEnvironmentClasses_ClassIDOnEnvironment(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Environments: []models.Environment{
			{EnvironmentID: "env:chrome", Name: "Chrome", ClassID: "envclass:browser"},
			{EnvironmentID: "env:safari", Name: "Safari", ClassID: "envclass:browser"},
		},
		EnvironmentClasses: []models.EnvironmentClass{
			{
				ClassID:   "envclass:browser",
				Name:      "Browsers",
				Dimension: "browser",
				// No MemberIDs — environments declare their class via ClassID.
			},
		},
	}

	g := Build(snap)

	outgoing := g.Outgoing("envclass:browser")
	containsEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeEnvironmentClassContains {
			containsEdges++
		}
	}
	if containsEdges != 2 {
		t.Errorf("expected 2 contains edges from ClassID field, got %d", containsEdges)
	}
}

func TestBuildEnvironmentClasses_NoDuplicateEdges(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Environments: []models.Environment{
			{EnvironmentID: "env:linux", Name: "Linux", ClassID: "envclass:os"},
		},
		EnvironmentClasses: []models.EnvironmentClass{
			{
				ClassID:   "envclass:os",
				Name:      "OS",
				MemberIDs: []string{"env:linux"}, // Also in MemberIDs.
			},
		},
	}

	g := Build(snap)

	outgoing := g.Outgoing("envclass:os")
	containsEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeEnvironmentClassContains {
			containsEdges++
		}
	}
	// Should be 1, not 2 — deduplication should prevent double-linking.
	if containsEdges != 1 {
		t.Errorf("expected 1 contains edge (no duplicates), got %d", containsEdges)
	}
}

func TestBuildDeviceConfigs_CreatesNodes(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		DeviceConfigs: []models.DeviceConfig{
			{
				DeviceID:      "device:iphone-15-ios17",
				Name:          "iPhone 15 iOS 17",
				Platform:      "ios",
				FormFactor:    "phone",
				OSVersion:     "17.0",
				Capabilities:  []string{"touch", "camera", "biometrics"},
				InferredFrom:  "xcode-scheme",
			},
			{
				DeviceID:      "device:chrome-120",
				Name:          "Chrome 120",
				Platform:      "web-browser",
				FormFactor:    "desktop",
				BrowserEngine: "chromium",
				InferredFrom:  "playwright-config",
			},
		},
	}

	g := Build(snap)

	iphone := g.Node("device:iphone-15-ios17")
	if iphone == nil {
		t.Fatal("expected device:iphone-15-ios17 node")
	}
	if iphone.Type != NodeDeviceConfig {
		t.Errorf("expected device_config type, got %s", iphone.Type)
	}
	if iphone.Metadata["platform"] != "ios" {
		t.Errorf("expected platform=ios, got %q", iphone.Metadata["platform"])
	}
	if iphone.Metadata["formFactor"] != "phone" {
		t.Errorf("expected formFactor=phone, got %q", iphone.Metadata["formFactor"])
	}

	chrome := g.Node("device:chrome-120")
	if chrome == nil {
		t.Fatal("expected device:chrome-120 node")
	}
	if chrome.Metadata["browserEngine"] != "chromium" {
		t.Errorf("expected browserEngine=chromium, got %q", chrome.Metadata["browserEngine"])
	}
}

func TestBuildDeviceConfigs_SkipsEmptyID(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		DeviceConfigs: []models.DeviceConfig{
			{DeviceID: "", Name: "missing"},
			{DeviceID: "device:valid", Name: "valid"},
		},
	}

	g := Build(snap)

	deviceNodes := g.NodesByType(NodeDeviceConfig)
	if len(deviceNodes) != 1 {
		t.Errorf("expected 1 device node, got %d", len(deviceNodes))
	}
}

func TestBuildDeviceConfigs_ConnectsToClass(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		EnvironmentClasses: []models.EnvironmentClass{
			{ClassID: "envclass:browser", Name: "Browsers", Dimension: "browser"},
		},
		DeviceConfigs: []models.DeviceConfig{
			{
				DeviceID:      "device:chrome-120",
				Name:          "Chrome 120",
				Platform:      "web-browser",
				BrowserEngine: "chromium",
				ClassID:       "envclass:browser",
			},
			{
				DeviceID:      "device:safari-17",
				Name:          "Safari 17",
				Platform:      "web-browser",
				BrowserEngine: "webkit",
				ClassID:       "envclass:browser",
			},
		},
	}

	g := Build(snap)

	outgoing := g.Outgoing("envclass:browser")
	containsEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeEnvironmentClassContains {
			containsEdges++
		}
	}
	if containsEdges != 2 {
		t.Errorf("expected 2 contains edges for browser class, got %d", containsEdges)
	}
}

func TestBuildEnvironments_EmptySnapshot(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}

	g := Build(snap)

	envNodes := g.NodesByType(NodeEnvironment)
	if len(envNodes) != 0 {
		t.Errorf("expected 0 environment nodes, got %d", len(envNodes))
	}
	classNodes := g.NodesByType(NodeEnvironmentClass)
	if len(classNodes) != 0 {
		t.Errorf("expected 0 environment class nodes, got %d", len(classNodes))
	}
	deviceNodes := g.NodesByType(NodeDeviceConfig)
	if len(deviceNodes) != 0 {
		t.Errorf("expected 0 device config nodes, got %d", len(deviceNodes))
	}
}

func TestBuildEnvironments_NodeFamily(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Environments: []models.Environment{
			{EnvironmentID: "env:test", Name: "Test"},
		},
		EnvironmentClasses: []models.EnvironmentClass{
			{ClassID: "envclass:test", Name: "Test Class"},
		},
		DeviceConfigs: []models.DeviceConfig{
			{DeviceID: "device:test", Name: "Test Device"},
		},
	}

	g := Build(snap)

	env := g.Node("env:test")
	if env.Family() != FamilyEnvironment {
		t.Errorf("expected environment family, got %s", env.Family())
	}
	class := g.Node("envclass:test")
	if class.Family() != FamilyEnvironment {
		t.Errorf("expected environment family for class, got %s", class.Family())
	}
	device := g.Node("device:test")
	if device.Family() != FamilyEnvironment {
		t.Errorf("expected environment family for device, got %s", device.Family())
	}
}

func TestBuildEnvironmentClasses_SkipsMissingMembers(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Environments: []models.Environment{
			{EnvironmentID: "env:exists", Name: "Exists"},
		},
		EnvironmentClasses: []models.EnvironmentClass{
			{
				ClassID:   "envclass:partial",
				Name:      "Partial",
				MemberIDs: []string{"env:exists", "env:missing"},
			},
		},
	}

	g := Build(snap)

	outgoing := g.Outgoing("envclass:partial")
	containsEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeEnvironmentClassContains {
			containsEdges++
		}
	}
	// Only env:exists should be linked, env:missing should be skipped.
	if containsEdges != 1 {
		t.Errorf("expected 1 contains edge (skipping missing member), got %d", containsEdges)
	}
}

func TestBuildEnvironmentEdges_TestFileToEnvironment(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/auth_test.go",
				Framework:      "go",
				TestCount:      3,
				EnvironmentIDs: []string{"env:ci-linux", "env:ci-macos"},
			},
		},
		Environments: []models.Environment{
			{EnvironmentID: "env:ci-linux", Name: "Linux"},
			{EnvironmentID: "env:ci-macos", Name: "macOS"},
		},
	}

	g := Build(snap)

	fileID := "file:test/auth_test.go"
	outgoing := g.Outgoing(fileID)
	envEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeTargetsEnvironment {
			envEdges++
			if e.Confidence != 0.8 {
				t.Errorf("expected confidence 0.8, got %f", e.Confidence)
			}
		}
	}
	if envEdges != 2 {
		t.Errorf("expected 2 environment edges, got %d", envEdges)
	}
}

func TestBuildEnvironmentEdges_TestFileToDevice(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:      "test/e2e/login.spec.ts",
				Framework: "playwright",
				TestCount: 5,
				DeviceIDs: []string{"device:chrome-120", "device:safari-17"},
			},
		},
		DeviceConfigs: []models.DeviceConfig{
			{DeviceID: "device:chrome-120", Name: "Chrome 120", Platform: "web-browser"},
			{DeviceID: "device:safari-17", Name: "Safari 17", Platform: "web-browser"},
		},
	}

	g := Build(snap)

	fileID := "file:test/e2e/login.spec.ts"
	outgoing := g.Outgoing(fileID)
	deviceEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeTargetsEnvironment {
			deviceEdges++
			if e.Confidence != 0.7 {
				t.Errorf("expected confidence 0.7 for device edges, got %f", e.Confidence)
			}
		}
	}
	if deviceEdges != 2 {
		t.Errorf("expected 2 device edges, got %d", deviceEdges)
	}
}

func TestBuildEnvironmentEdges_ScenarioToEnvironment(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Scenarios: []models.Scenario{
			{
				ScenarioID:     "scenario:auth:login-flow",
				Name:           "Login Flow",
				Framework:      "deepeval",
				EnvironmentIDs: []string{"env:staging"},
				Executable:     true,
			},
		},
		Environments: []models.Environment{
			{EnvironmentID: "env:staging", Name: "Staging", IsProduction: false},
		},
	}

	g := Build(snap)

	outgoing := g.Outgoing("scenario:auth:login-flow")
	envEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeTargetsEnvironment && e.To == "env:staging" {
			envEdges++
		}
	}
	if envEdges != 1 {
		t.Errorf("expected 1 environment edge from scenario, got %d", envEdges)
	}
}

func TestBuildEnvironmentEdges_SkipsMissingTargets(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/a.test.js",
				Framework:      "jest",
				TestCount:      1,
				EnvironmentIDs: []string{"env:nonexistent"},
				DeviceIDs:      []string{"device:nonexistent"},
			},
		},
		// No environments or devices defined — IDs don't resolve.
	}

	g := Build(snap)

	fileID := "file:test/a.test.js"
	outgoing := g.Outgoing(fileID)
	envEdges := 0
	for _, e := range outgoing {
		if e.Type == EdgeTargetsEnvironment {
			envEdges++
		}
	}
	if envEdges != 0 {
		t.Errorf("expected 0 environment edges for missing targets, got %d", envEdges)
	}
}

func TestBuildEnvironmentEdges_FullTraversal(t *testing.T) {
	t.Parallel()
	// End-to-end: test file → environment → class. Verify the full chain
	// is traversable via graph edges.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/payment.test.js",
				Framework:      "jest",
				TestCount:      2,
				EnvironmentIDs: []string{"env:ci-linux"},
				DeviceIDs:      []string{"device:chrome-120"},
			},
		},
		Environments: []models.Environment{
			{EnvironmentID: "env:ci-linux", Name: "Linux", OS: "linux"},
		},
		EnvironmentClasses: []models.EnvironmentClass{
			{
				ClassID:   "envclass:os",
				Name:      "Operating Systems",
				Dimension: "os",
				MemberIDs: []string{"env:ci-linux"},
			},
			{
				ClassID:   "envclass:browser",
				Name:      "Browsers",
				Dimension: "browser",
			},
		},
		DeviceConfigs: []models.DeviceConfig{
			{
				DeviceID:      "device:chrome-120",
				Name:          "Chrome 120",
				Platform:      "web-browser",
				BrowserEngine: "chromium",
				ClassID:       "envclass:browser",
			},
		},
	}

	g := Build(snap)

	// TestFile → env:ci-linux
	fileID := "file:test/payment.test.js"
	foundEnv := false
	foundDevice := false
	for _, e := range g.Outgoing(fileID) {
		if e.Type == EdgeTargetsEnvironment && e.To == "env:ci-linux" {
			foundEnv = true
		}
		if e.Type == EdgeTargetsEnvironment && e.To == "device:chrome-120" {
			foundDevice = true
		}
	}
	if !foundEnv {
		t.Error("expected TestFile → env:ci-linux edge")
	}
	if !foundDevice {
		t.Error("expected TestFile → device:chrome-120 edge")
	}

	// envclass:os → env:ci-linux
	foundClassToEnv := false
	for _, e := range g.Outgoing("envclass:os") {
		if e.Type == EdgeEnvironmentClassContains && e.To == "env:ci-linux" {
			foundClassToEnv = true
		}
	}
	if !foundClassToEnv {
		t.Error("expected envclass:os → env:ci-linux edge")
	}

	// envclass:browser → device:chrome-120
	foundClassToDevice := false
	for _, e := range g.Outgoing("envclass:browser") {
		if e.Type == EdgeEnvironmentClassContains && e.To == "device:chrome-120" {
			foundClassToDevice = true
		}
	}
	if !foundClassToDevice {
		t.Error("expected envclass:browser → device:chrome-120 edge")
	}
}
