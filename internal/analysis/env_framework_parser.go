package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
	"gopkg.in/yaml.v3"
)

// FrameworkMatrixResult holds device configs and environment info from test
// framework configuration parsing.
type FrameworkMatrixResult struct {
	DeviceConfigs      []models.DeviceConfig
	Environments       []models.Environment
	EnvironmentClasses []models.EnvironmentClass
}

// ParseFrameworkMatrices scans repository configuration files and test files
// for test framework matrix definitions — Playwright browser configs, mobile
// device matrices, and pytest parametrize markers.
func ParseFrameworkMatrices(root string, testFiles []models.TestFile) *FrameworkMatrixResult {
	result := &FrameworkMatrixResult{}

	parsePlaywrightConfig(root, result)
	parsePytestIni(root, result)
	parsePytestParametrize(root, testFiles, result)
	parseMobileDeviceConfigs(root, result)

	return result
}

// --- Playwright browser configuration ---

// playwrightConfig is a minimal representation of playwright.config.ts/js.
// We parse it with regex since it's JS/TS, not YAML.
var (
	// Match projects: [ { name: 'chromium', use: { ...devices['Desktop Chrome'] } }, ... ]
	pwProjectNamePattern = regexp.MustCompile(`name:\s*['"](\w[\w\s-]*)['"]`)
	// Match use: { browserName: 'chromium' }
	pwBrowserNamePattern = regexp.MustCompile(`browserName:\s*['"](\w+)['"]`)
	// Match ...devices['iPhone 13']
	pwDevicePattern = regexp.MustCompile(`devices\[['"]([^'"]+)['"]\]`)
)

func parsePlaywrightConfig(root string, result *FrameworkMatrixResult) {
	configPaths := []string{
		"playwright.config.ts",
		"playwright.config.js",
		"playwright.config.mjs",
	}

	var content string
	for _, p := range configPaths {
		data, err := os.ReadFile(filepath.Join(root, p))
		if err == nil {
			content = string(data)
			break
		}
	}
	if content == "" {
		return
	}

	provenance := "playwright-config"

	// Extract project names (browser names).
	browsers := map[string]bool{}
	for _, m := range pwProjectNamePattern.FindAllStringSubmatch(content, -1) {
		name := strings.TrimSpace(m[1])
		if isPWBrowser(name) {
			browsers[name] = true
		}
	}
	for _, m := range pwBrowserNamePattern.FindAllStringSubmatch(content, -1) {
		browsers[m[1]] = true
	}

	// Extract device references.
	devices := map[string]bool{}
	for _, m := range pwDevicePattern.FindAllStringSubmatch(content, -1) {
		devices[m[1]] = true
	}

	// Create browser environment class + environments.
	if len(browsers) > 0 {
		classID := "envclass:pw-browser"
		memberIDs := make([]string, 0, len(browsers))
		sortedBrowsers := sortedStringSet(browsers)

		for _, browser := range sortedBrowsers {
			envID := "env:pw-" + sanitizeID(browser)
			memberIDs = append(memberIDs, envID)

			result.Environments = append(result.Environments, models.Environment{
				EnvironmentID: envID,
				Name:          browser,
				CIProvider:    "",
				ClassID:       classID,
				InferredFrom:  provenance,
			})
		}

		result.EnvironmentClasses = append(result.EnvironmentClasses, models.EnvironmentClass{
			ClassID:   classID,
			Name:      "Playwright browsers",
			Dimension: "browser",
			MemberIDs: memberIDs,
		})
	}

	// Create device configs from Playwright device references.
	if len(devices) > 0 {
		classID := "envclass:pw-device"
		memberIDs := make([]string, 0, len(devices))
		sortedDevices := sortedStringSet(devices)

		for _, device := range sortedDevices {
			deviceID := "device:pw-" + sanitizeID(device)
			memberIDs = append(memberIDs, deviceID)

			result.DeviceConfigs = append(result.DeviceConfigs, models.DeviceConfig{
				DeviceID:     deviceID,
				Name:         device,
				Platform:     inferDevicePlatform(device),
				FormFactor:   inferFormFactor(device),
				ClassID:      classID,
				InferredFrom: provenance,
			})
		}

		result.EnvironmentClasses = append(result.EnvironmentClasses, models.EnvironmentClass{
			ClassID:   classID,
			Name:      "Playwright devices",
			Dimension: "device",
			MemberIDs: memberIDs,
		})
	}
}

func isPWBrowser(name string) bool {
	lower := strings.ToLower(name)
	return lower == "chromium" || lower == "firefox" || lower == "webkit" ||
		lower == "chrome" || lower == "msedge" || lower == "safari" ||
		strings.Contains(lower, "mobile chrome") || strings.Contains(lower, "mobile safari")
}

// --- pytest configuration and parametrize ---

func parsePytestIni(root string, result *FrameworkMatrixResult) {
	// Check pytest.ini, pyproject.toml, setup.cfg for markers or environment config.
	// This is a lightweight check for environment-related markers.
	configPaths := []string{
		filepath.Join(root, "pytest.ini"),
		filepath.Join(root, "pyproject.toml"),
	}

	for _, p := range configPaths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		content := string(data)

		// Look for env markers in markers section.
		if strings.Contains(content, "[tool.pytest") || strings.Contains(content, "[pytest]") {
			extractPytestEnvMarkers(content, result)
		}
	}
}

var pytestMarkerPattern = regexp.MustCompile(`(?m)^\s*(\w+):\s*(?:mark\s+for\s+)?(?:run(?:ning)?\s+(?:on|in|with)\s+)?(.+)$`)

func extractPytestEnvMarkers(content string, result *FrameworkMatrixResult) {
	// Look for marker patterns like: linux: run on Linux, browser_chrome: Chrome browser tests
	envMarkers := map[string]string{} // marker → description
	lines := strings.Split(content, "\n")
	inMarkers := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "markers") && strings.Contains(trimmed, "=") {
			inMarkers = true
			continue
		}
		if inMarkers {
			if trimmed == "" || (!strings.HasPrefix(trimmed, " ") && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(trimmed, "\"") && trimmed != "" && !strings.Contains(trimmed, ":")) {
				inMarkers = false
				continue
			}
			// Clean up TOML multiline string markers.
			cleaned := strings.Trim(trimmed, "\"',")
			if m := pytestMarkerPattern.FindStringSubmatch(cleaned); m != nil {
				marker := m[1]
				if isEnvironmentMarker(marker) {
					envMarkers[marker] = strings.TrimSpace(m[2])
				}
			}
		}
	}

	// We don't create environments from markers alone — they're informational.
	// They become relevant when pytest.parametrize or conftest fixtures reference them.
	_ = envMarkers
}

func isEnvironmentMarker(marker string) bool {
	lower := strings.ToLower(marker)
	return strings.Contains(lower, "linux") || strings.Contains(lower, "macos") ||
		strings.Contains(lower, "windows") || strings.Contains(lower, "chrome") ||
		strings.Contains(lower, "firefox") || strings.Contains(lower, "safari") ||
		strings.Contains(lower, "browser") || strings.Contains(lower, "mobile") ||
		strings.Contains(lower, "device")
}

// parsePytestParametrize scans Python test files for @pytest.mark.parametrize
// decorators that define environment or browser matrices.
var (
	pytestParametrizePattern = regexp.MustCompile(`@pytest\.mark\.parametrize\s*\(\s*['"](\w+)['"]\s*,\s*\[([^\]]+)\]`)
	pytestParamValuePattern  = regexp.MustCompile(`['"]([^'"]+)['"]`)
)

func parsePytestParametrize(root string, testFiles []models.TestFile, result *FrameworkMatrixResult) {
	for _, tf := range testFiles {
		if !strings.HasSuffix(tf.Path, ".py") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, tf.Path))
		if err != nil {
			continue
		}
		src := string(data)

		matches := pytestParametrizePattern.FindAllStringSubmatch(src, -1)
		for _, m := range matches {
			paramName := m[1]
			paramValues := m[2]

			if !isEnvironmentParam(paramName) {
				continue
			}

			values := pytestParamValuePattern.FindAllStringSubmatch(paramValues, -1)
			if len(values) == 0 {
				continue
			}

			dimension := inferDimension(paramName)
			classID := "envclass:pytest-" + sanitizeID(paramName)
			memberIDs := make([]string, 0, len(values))

			for _, v := range values {
				val := v[1]
				envID := "env:pytest-" + sanitizeID(paramName) + "-" + sanitizeID(val)
				memberIDs = append(memberIDs, envID)

				result.Environments = appendEnvIfNew(result.Environments, models.Environment{
					EnvironmentID: envID,
					Name:          paramName + " " + val,
					ClassID:       classID,
					InferredFrom:  "pytest-parametrize:" + tf.Path,
				})
			}

			result.EnvironmentClasses = appendClassIfNew(result.EnvironmentClasses, models.EnvironmentClass{
				ClassID:   classID,
				Name:      paramName,
				Dimension: dimension,
				MemberIDs: memberIDs,
			})
		}
	}
}

func isEnvironmentParam(name string) bool {
	lower := strings.ToLower(name)
	return lower == "browser" || lower == "browsers" ||
		lower == "os" || lower == "platform" ||
		lower == "device" || lower == "devices" ||
		lower == "runtime" || lower == "version" ||
		lower == "env" || lower == "environment" ||
		lower == "python_version" || lower == "node_version"
}

// --- Mobile device matrices ---

// parseMobileDeviceConfigs looks for device matrix definitions in config files
// commonly used by mobile test frameworks.
func parseMobileDeviceConfigs(root string, result *FrameworkMatrixResult) {
	// BrowserStack config (browserstack.json or .browserstack.yml)
	parseBrowserStackConfig(root, result)
	// Appium capabilities in JSON
	parseAppiumConfig(root, result)
}

func parseBrowserStackConfig(root string, result *FrameworkMatrixResult) {
	paths := []string{
		filepath.Join(root, "browserstack.json"),
		filepath.Join(root, ".browserstack.yml"),
		filepath.Join(root, "browserstack.yml"),
	}

	var data []byte
	var isYAML bool
	for _, p := range paths {
		d, err := os.ReadFile(p)
		if err == nil {
			data = d
			isYAML = strings.HasSuffix(p, ".yml") || strings.HasSuffix(p, ".yaml")
			break
		}
	}
	if data == nil {
		return
	}

	provenance := "browserstack"

	var raw map[string]interface{}
	if isYAML {
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return
		}
	} else {
		// JSON is valid YAML.
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return
		}
	}

	// Extract platforms array: [{os, os_version, browser, browser_version, device}, ...]
	platforms := extractInterfaceSlice(raw, "platforms")
	if platforms == nil {
		platforms = extractInterfaceSlice(raw, "browsers")
	}

	classID := "envclass:browserstack-device"
	var memberIDs []string

	for _, p := range platforms {
		pMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		device := stringFromMap(pMap, "device")
		browser := stringFromMap(pMap, "browser")
		osName := stringFromMap(pMap, "os")
		osVersion := stringFromMap(pMap, "os_version")

		var name string
		if device != "" {
			name = device
		} else if browser != "" {
			name = browser
			if osName != "" {
				name += " on " + osName
			}
		} else {
			continue
		}

		deviceID := "device:bs-" + sanitizeID(name)
		memberIDs = append(memberIDs, deviceID)

		dc := models.DeviceConfig{
			DeviceID:     deviceID,
			Name:         name,
			Platform:     inferDevicePlatform(name),
			FormFactor:   inferFormFactor(name),
			OSVersion:    osVersion,
			ClassID:      classID,
			InferredFrom: provenance,
		}
		if browser != "" {
			dc.BrowserEngine = inferBrowserEngine(browser)
		}

		result.DeviceConfigs = appendDeviceIfNew(result.DeviceConfigs, dc)
	}

	if len(memberIDs) > 0 {
		result.EnvironmentClasses = appendClassIfNew(result.EnvironmentClasses, models.EnvironmentClass{
			ClassID:   classID,
			Name:      "BrowserStack devices",
			Dimension: "device",
			MemberIDs: memberIDs,
		})
	}
}

func parseAppiumConfig(root string, result *FrameworkMatrixResult) {
	paths := []string{
		filepath.Join(root, "appium.conf.json"),
		filepath.Join(root, "wdio.conf.js"),
		filepath.Join(root, "wdio.conf.ts"),
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		content := string(data)
		extractAppiumDevices(content, filepath.Base(p), result)
	}
}

var (
	appiumDeviceNamePattern  = regexp.MustCompile(`['"]appium:deviceName['"]\s*:\s*['"]([^'"]+)['"]`)
	appiumPlatformPattern    = regexp.MustCompile(`['"](?:appium:)?platformName['"]\s*:\s*['"]([^'"]+)['"]`)
	appiumPlatformVerPattern = regexp.MustCompile(`['"](?:appium:)?platformVersion['"]\s*:\s*['"]([^'"]+)['"]`)
)

func extractAppiumDevices(content, filename string, result *FrameworkMatrixResult) {
	deviceNames := appiumDeviceNamePattern.FindAllStringSubmatch(content, -1)
	if len(deviceNames) == 0 {
		return
	}

	provenance := "appium:" + filename
	classID := "envclass:appium-device"
	var memberIDs []string

	platforms := appiumPlatformPattern.FindAllStringSubmatch(content, -1)
	versions := appiumPlatformVerPattern.FindAllStringSubmatch(content, -1)

	for i, m := range deviceNames {
		name := m[1]
		deviceID := "device:appium-" + sanitizeID(name)
		memberIDs = append(memberIDs, deviceID)

		dc := models.DeviceConfig{
			DeviceID:     deviceID,
			Name:         name,
			Platform:     inferDevicePlatform(name),
			FormFactor:   inferFormFactor(name),
			ClassID:      classID,
			InferredFrom: provenance,
		}

		if i < len(platforms) {
			dc.Platform = strings.ToLower(platforms[i][1])
		}
		if i < len(versions) {
			dc.OSVersion = versions[i][1]
		}

		result.DeviceConfigs = appendDeviceIfNew(result.DeviceConfigs, dc)
	}

	if len(memberIDs) > 0 {
		result.EnvironmentClasses = appendClassIfNew(result.EnvironmentClasses, models.EnvironmentClass{
			ClassID:   classID,
			Name:      "Appium devices",
			Dimension: "device",
			MemberIDs: memberIDs,
		})
	}
}

// --- Helpers ---

func inferDevicePlatform(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ios"):
		return "ios"
	case strings.Contains(lower, "pixel") || strings.Contains(lower, "galaxy") || strings.Contains(lower, "android"):
		return "android"
	case strings.Contains(lower, "chrome") || strings.Contains(lower, "firefox") ||
		strings.Contains(lower, "safari") || strings.Contains(lower, "edge") ||
		strings.Contains(lower, "webkit") || strings.Contains(lower, "desktop"):
		return "web-browser"
	default:
		return ""
	}
}

func inferFormFactor(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet") || strings.Contains(lower, "tab "):
		return "tablet"
	case strings.Contains(lower, "desktop"):
		return "desktop"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "pixel") ||
		strings.Contains(lower, "galaxy") || strings.Contains(lower, "phone") ||
		strings.Contains(lower, "mobile"):
		return "phone"
	default:
		return ""
	}
}

func inferBrowserEngine(browser string) string {
	lower := strings.ToLower(browser)
	switch {
	case strings.Contains(lower, "chrome") || strings.Contains(lower, "chromium") || strings.Contains(lower, "edge"):
		return "chromium"
	case strings.Contains(lower, "firefox"):
		return "gecko"
	case strings.Contains(lower, "safari") || strings.Contains(lower, "webkit"):
		return "webkit"
	default:
		return ""
	}
}

func extractInterfaceSlice(m map[string]interface{}, key string) []interface{} {
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return arr
}

func stringFromMap(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func appendDeviceIfNew(devices []models.DeviceConfig, dc models.DeviceConfig) []models.DeviceConfig {
	for _, d := range devices {
		if d.DeviceID == dc.DeviceID {
			return devices
		}
	}
	return append(devices, dc)
}

func sortedStringSet(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
