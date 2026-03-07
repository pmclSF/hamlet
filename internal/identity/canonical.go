package identity

import "strings"

// Delimiter separates components of a canonical identity string.
const Delimiter = "::"

// BuildCanonical constructs a canonical identity string from its components.
//
// Format: {normalized_path}::{suite_hierarchy}::{test_name}
//
// If paramSignature is non-empty, it is appended:
//   {normalized_path}::{suite_hierarchy}::{test_name}::{param_signature}
//
// All components are normalized before joining.
func BuildCanonical(filePath string, suiteHierarchy []string, testName string, paramSignature string) string {
	var parts []string

	parts = append(parts, NormalizePath(filePath))
	parts = append(parts, NormalizeSuiteHierarchy(suiteHierarchy))
	parts = append(parts, NormalizeName(testName))

	if paramSignature != "" {
		parts = append(parts, NormalizeName(paramSignature))
	}

	return strings.Join(parts, Delimiter)
}

// ParseCanonical splits a canonical identity back into its components.
// Returns (path, suiteHierarchy, testName, paramSignature).
func ParseCanonical(canonical string) (path string, suiteHierarchy string, testName string, paramSignature string) {
	parts := strings.SplitN(canonical, Delimiter, 4)
	switch len(parts) {
	case 4:
		return parts[0], parts[1], parts[2], parts[3]
	case 3:
		return parts[0], parts[1], parts[2], ""
	case 2:
		return parts[0], "", parts[1], ""
	case 1:
		return parts[0], "", "", ""
	default:
		return "", "", "", ""
	}
}
