package models

// Environment represents a concrete execution context where tests run.
//
// An environment is a specific, observed instance: "GitHub Actions Ubuntu 22.04
// with Node 22" or "staging deployment on AWS us-east-1". Environments are
// inferred from CI configuration files (workflow matrices, runs-on directives,
// Docker configs) or declared explicitly in terrain.yaml.
//
// Environments are graph nodes (NodeEnvironment) connected to test files and
// executions via EdgeTargetsEnvironment edges. Coverage is per-environment:
// a test passing on Linux does not automatically cover macOS.
type Environment struct {
	// EnvironmentID is a stable identifier.
	// Format: "env:<canonical-name>".
	// Examples: "env:ci-linux-node22", "env:staging", "env:local-macos".
	EnvironmentID string `json:"environmentId"`

	// Name is a human-readable label.
	Name string `json:"name"`

	// OS is the operating system (linux, macos, windows).
	OS string `json:"os,omitempty"`

	// OSVersion is the OS version if known (22.04, 14.2, etc.).
	OSVersion string `json:"osVersion,omitempty"`

	// Runtime is the language runtime (node-22, go-1.22, python-3.12).
	Runtime string `json:"runtime,omitempty"`

	// CIProvider is the CI system (github-actions, gitlab-ci, jenkins).
	CIProvider string `json:"ciProvider,omitempty"`

	// ResourceClass is the compute tier if known (large, xlarge, gpu).
	ResourceClass string `json:"resourceClass,omitempty"`

	// IsProduction indicates whether this is a production-like environment.
	IsProduction bool `json:"isProduction,omitempty"`

	// ClassID links this concrete environment to its EnvironmentClass, if any.
	ClassID string `json:"classId,omitempty"`

	// InferredFrom identifies the source of environment information.
	// Examples: "github-actions:matrix", "dockerfile", "manual".
	InferredFrom string `json:"inferredFrom,omitempty"`
}

// EnvironmentClass represents a group of related environments that share
// common characteristics. Classes are the mechanism for environment matrices:
// a "browser" class might contain Chrome, Safari, and Firefox environments;
// a "platform" class might contain Linux, macOS, and Windows.
//
// EnvironmentClasses are graph nodes (NodeEnvironmentClass) connected to
// their member environments via EdgeEnvironmentClassContains edges.
// They enable coverage questions like "do we test on all supported browsers?"
type EnvironmentClass struct {
	// ClassID is a stable identifier.
	// Format: "envclass:<name>".
	// Examples: "envclass:browser", "envclass:os", "envclass:runtime".
	ClassID string `json:"classId"`

	// Name is a human-readable label for this class.
	Name string `json:"name"`

	// Dimension describes what this class varies on.
	// Values: "os", "runtime", "browser", "device", "region", "provider".
	Dimension string `json:"dimension,omitempty"`

	// MemberIDs lists the EnvironmentIDs belonging to this class.
	MemberIDs []string `json:"memberIds,omitempty"`
}

// DeviceConfig represents a target device or browser where tests execute.
//
// DeviceConfigs model the device dimension of test execution: which phones,
// tablets, browsers, or emulators tests target. They are distinct from
// environments (which model the CI/infra dimension) but connected: a test
// might run in environment "env:ci-linux" targeting device "device:iphone-15".
//
// DeviceConfigs are graph nodes (NodeDeviceConfig) and connect to tests
// and environments. They enable mobile/device matrix analysis and
// cross-browser coverage tracking.
type DeviceConfig struct {
	// DeviceID is a stable identifier.
	// Format: "device:<canonical-name>".
	// Examples: "device:iphone-15-ios17", "device:chrome-120", "device:pixel-8".
	DeviceID string `json:"deviceId"`

	// Name is a human-readable label.
	Name string `json:"name"`

	// Platform is the platform category (ios, android, web-browser).
	Platform string `json:"platform,omitempty"`

	// FormFactor is the physical form (phone, tablet, desktop).
	FormFactor string `json:"formFactor,omitempty"`

	// OSVersion is the device OS version.
	OSVersion string `json:"osVersion,omitempty"`

	// BrowserEngine is the rendering engine for web browsers
	// (chromium, webkit, gecko).
	BrowserEngine string `json:"browserEngine,omitempty"`

	// Capabilities lists device-specific capabilities
	// (touch, camera, biometrics, nfc).
	Capabilities []string `json:"capabilities,omitempty"`

	// ClassID links this device to an EnvironmentClass, if any.
	ClassID string `json:"classId,omitempty"`

	// InferredFrom identifies the source of device information.
	// Examples: "playwright-config", "browserstack", "xcode-scheme".
	InferredFrom string `json:"inferredFrom,omitempty"`
}
