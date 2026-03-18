package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestParsePlaywrightConfig_Browsers(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "playwright.config.ts"), []byte(`
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'webkit', use: { ...devices['Desktop Safari'] } },
    { name: 'Mobile Chrome', use: { ...devices['Pixel 5'] } },
    { name: 'Mobile Safari', use: { ...devices['iPhone 13'] } },
  ],
});
`), 0o644)

	result := &FrameworkMatrixResult{}
	parsePlaywrightConfig(root, result)

	// Should detect browser class with chromium, firefox, webkit, Mobile Chrome, Mobile Safari.
	browserClassFound := false
	for _, cls := range result.EnvironmentClasses {
		if cls.Dimension == "browser" {
			browserClassFound = true
			if len(cls.MemberIDs) < 3 {
				t.Errorf("browser class: want at least 3 members, got %d", len(cls.MemberIDs))
			}
		}
	}
	if !browserClassFound {
		t.Error("expected browser environment class from Playwright config")
	}

	// Should detect device class with Pixel 5, iPhone 13, Desktop Chrome, etc.
	deviceClassFound := false
	for _, cls := range result.EnvironmentClasses {
		if cls.Dimension == "device" {
			deviceClassFound = true
			if len(cls.MemberIDs) < 2 {
				t.Errorf("device class: want at least 2 members, got %d", len(cls.MemberIDs))
			}
		}
	}
	if !deviceClassFound {
		t.Error("expected device environment class from Playwright config")
	}

	// Check device configs.
	devicePlatforms := map[string]string{} // name → platform
	for _, dc := range result.DeviceConfigs {
		devicePlatforms[dc.Name] = dc.Platform
	}
	if p, ok := devicePlatforms["iPhone 13"]; !ok {
		t.Error("expected iPhone 13 device config")
	} else if p != "ios" {
		t.Errorf("iPhone 13: want platform ios, got %s", p)
	}
	if p, ok := devicePlatforms["Pixel 5"]; !ok {
		t.Error("expected Pixel 5 device config")
	} else if p != "android" {
		t.Errorf("Pixel 5: want platform android, got %s", p)
	}
}

func TestParsePlaywrightConfig_BrowserName(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "playwright.config.js"), []byte(`
module.exports = {
  projects: [
    { use: { browserName: 'chromium' } },
    { use: { browserName: 'firefox' } },
  ],
};
`), 0o644)

	result := &FrameworkMatrixResult{}
	parsePlaywrightConfig(root, result)

	if len(result.Environments) < 2 {
		t.Fatalf("expected at least 2 browser environments, got %d", len(result.Environments))
	}

	names := map[string]bool{}
	for _, env := range result.Environments {
		names[env.Name] = true
	}
	if !names["chromium"] {
		t.Error("expected chromium environment")
	}
	if !names["firefox"] {
		t.Error("expected firefox environment")
	}
}

func TestParsePytestParametrize_BrowserMatrix(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.MkdirAll(filepath.Join(root, "tests"), 0o755)
	os.WriteFile(filepath.Join(root, "tests", "test_cross_browser.py"), []byte(`
import pytest

@pytest.mark.parametrize("browser", ["chrome", "firefox", "safari", "edge"])
def test_login(browser):
    driver = get_driver(browser)
    driver.get("/login")
    assert driver.title == "Login"

@pytest.mark.parametrize("platform", ["linux", "macos", "windows"])
def test_platform(platform):
    assert True
`), 0o644)

	testFiles := []models.TestFile{
		{Path: "tests/test_cross_browser.py", Framework: "pytest"},
	}

	result := &FrameworkMatrixResult{}
	parsePytestParametrize(root, testFiles, result)

	classIDs := map[string]bool{}
	for _, cls := range result.EnvironmentClasses {
		classIDs[cls.ClassID] = true
	}

	if !classIDs["envclass:pytest-browser"] {
		t.Error("expected pytest browser class")
	}
	if !classIDs["envclass:pytest-platform"] {
		t.Error("expected pytest platform class")
	}

	// Check browser class members.
	for _, cls := range result.EnvironmentClasses {
		if cls.ClassID == "envclass:pytest-browser" {
			if len(cls.MemberIDs) != 4 {
				t.Errorf("browser class: want 4 members, got %d", len(cls.MemberIDs))
			}
			if cls.Dimension != "browser" {
				t.Errorf("browser class: want dimension browser, got %s", cls.Dimension)
			}
		}
		if cls.ClassID == "envclass:pytest-platform" {
			if len(cls.MemberIDs) != 3 {
				t.Errorf("platform class: want 3 members, got %d", len(cls.MemberIDs))
			}
			if cls.Dimension != "os" {
				t.Errorf("platform class: want dimension os, got %s", cls.Dimension)
			}
		}
	}

	// Verify provenance.
	for _, env := range result.Environments {
		if env.InferredFrom == "" {
			t.Errorf("env %s missing InferredFrom", env.EnvironmentID)
		}
	}
}

func TestParsePytestParametrize_NonEnvParam(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.MkdirAll(filepath.Join(root, "tests"), 0o755)
	os.WriteFile(filepath.Join(root, "tests", "test_math.py"), []byte(`
import pytest

@pytest.mark.parametrize("x,y,expected", [(1,2,3), (4,5,9)])
def test_add(x, y, expected):
    assert x + y == expected
`), 0o644)

	testFiles := []models.TestFile{
		{Path: "tests/test_math.py", Framework: "pytest"},
	}

	result := &FrameworkMatrixResult{}
	parsePytestParametrize(root, testFiles, result)

	// Non-environment params should be ignored.
	if len(result.EnvironmentClasses) != 0 {
		t.Errorf("expected 0 classes for non-env params, got %d", len(result.EnvironmentClasses))
	}
}

func TestParseBrowserStackConfig(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "browserstack.json"), []byte(`{
  "platforms": [
    {"os": "Windows", "os_version": "11", "browser": "Chrome", "browser_version": "120"},
    {"os": "OS X", "os_version": "Sonoma", "browser": "Safari", "browser_version": "17"},
    {"device": "iPhone 15", "os_version": "17"},
    {"device": "Samsung Galaxy S24", "os_version": "14"}
  ]
}`), 0o644)

	result := &FrameworkMatrixResult{}
	parseBrowserStackConfig(root, result)

	if len(result.DeviceConfigs) != 4 {
		t.Fatalf("expected 4 device configs, got %d", len(result.DeviceConfigs))
	}

	platforms := map[string]string{}
	for _, dc := range result.DeviceConfigs {
		platforms[dc.Name] = dc.Platform
	}

	if _, ok := platforms["iPhone 15"]; !ok {
		t.Error("expected iPhone 15 device")
	}
	if _, ok := platforms["Samsung Galaxy S24"]; !ok {
		t.Error("expected Samsung Galaxy S24 device")
	}

	// Browser engine detection.
	for _, dc := range result.DeviceConfigs {
		if dc.Name == "Chrome on Windows" || dc.Name == "Chrome" {
			if dc.BrowserEngine != "chromium" {
				t.Errorf("Chrome: want engine chromium, got %s", dc.BrowserEngine)
			}
		}
	}
}

func TestParseAppiumConfig(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "appium.conf.json"), []byte(`{
  "capabilities": [
    {
      "appium:deviceName": "Pixel 7",
      "platformName": "Android",
      "platformVersion": "13"
    },
    {
      "appium:deviceName": "iPhone 14 Pro",
      "platformName": "iOS",
      "platformVersion": "16.4"
    }
  ]
}`), 0o644)

	result := &FrameworkMatrixResult{}
	parseAppiumConfig(root, result)

	if len(result.DeviceConfigs) != 2 {
		t.Fatalf("expected 2 device configs, got %d", len(result.DeviceConfigs))
	}

	names := map[string]bool{}
	for _, dc := range result.DeviceConfigs {
		names[dc.Name] = true
	}
	if !names["Pixel 7"] {
		t.Error("expected Pixel 7 device")
	}
	if !names["iPhone 14 Pro"] {
		t.Error("expected iPhone 14 Pro device")
	}
}

func TestInferDevicePlatform(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want string
	}{
		{"iPhone 15 Pro", "ios"},
		{"iPad Air", "ios"},
		{"Pixel 8", "android"},
		{"Galaxy S24", "android"},
		{"Desktop Chrome", "web-browser"},
		{"Desktop Safari", "web-browser"},
		{"Unknown Device", ""},
	}
	for _, tc := range cases {
		got := inferDevicePlatform(tc.name)
		if got != tc.want {
			t.Errorf("inferDevicePlatform(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestInferFormFactor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want string
	}{
		{"iPhone 15", "phone"},
		{"iPad Air", "tablet"},
		{"Pixel 8", "phone"},
		{"Desktop Chrome", "desktop"},
		{"Galaxy Tab S9", "tablet"},
		{"Unknown", ""},
	}
	for _, tc := range cases {
		got := inferFormFactor(tc.name)
		if got != tc.want {
			t.Errorf("inferFormFactor(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestInferBrowserEngine(t *testing.T) {
	t.Parallel()
	cases := []struct {
		browser string
		want    string
	}{
		{"Chrome", "chromium"},
		{"chromium", "chromium"},
		{"Edge", "chromium"},
		{"Firefox", "gecko"},
		{"Safari", "webkit"},
		{"webkit", "webkit"},
		{"Unknown", ""},
	}
	for _, tc := range cases {
		got := inferBrowserEngine(tc.browser)
		if got != tc.want {
			t.Errorf("inferBrowserEngine(%q) = %q, want %q", tc.browser, got, tc.want)
		}
	}
}

func TestParseFrameworkMatrices_NoConfigs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	result := ParseFrameworkMatrices(root, nil)
	if len(result.DeviceConfigs) != 0 {
		t.Errorf("expected 0 device configs, got %d", len(result.DeviceConfigs))
	}
	if len(result.Environments) != 0 {
		t.Errorf("expected 0 environments, got %d", len(result.Environments))
	}
}

func TestParsePlaywrightConfig_DeviceFormFactors(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	os.WriteFile(filepath.Join(root, "playwright.config.ts"), []byte(`
export default {
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'Mobile Safari', use: { ...devices['iPhone 13'] } },
    { name: 'Tablet', use: { ...devices['iPad Pro 11'] } },
  ],
};
`), 0o644)

	result := &FrameworkMatrixResult{}
	parsePlaywrightConfig(root, result)

	formFactors := map[string]string{}
	for _, dc := range result.DeviceConfigs {
		formFactors[dc.Name] = dc.FormFactor
	}

	if ff, ok := formFactors["iPhone 13"]; !ok {
		t.Error("expected iPhone 13")
	} else if ff != "phone" {
		t.Errorf("iPhone 13: want form_factor phone, got %s", ff)
	}

	if ff, ok := formFactors["iPad Pro 11"]; !ok {
		t.Error("expected iPad Pro 11")
	} else if ff != "tablet" {
		t.Errorf("iPad Pro 11: want form_factor tablet, got %s", ff)
	}
}
