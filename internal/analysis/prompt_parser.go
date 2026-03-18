package analysis

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// ParseEmbeddedPrompts performs structured detection of AI prompts and
// contexts embedded in source code. Unlike flat regex, this parser examines
// multi-line structural patterns:
//
//   - Message arrays with role/content structure
//   - Template literals / f-strings containing AI instruction markers
//   - Few-shot example arrays
//   - Prompt builder functions (functions returning AI-instructed strings)
//   - Triple-quote / multi-line string prompts
//
// Each detection carries TierContent or TierSemantic with a confidence score.
// Non-AI strings are filtered by requiring corroborating AI markers.
func ParseEmbeddedPrompts(relPath, src, lang string) []models.CodeSurface {
	switch lang {
	case "js":
		return parseJSPrompts(relPath, src)
	case "python":
		return parsePythonPrompts(relPath, src)
	case "go":
		return parseGoPrompts(relPath, src)
	default:
		return nil
	}
}

// --- JS/TS structured parsing ---

// AI markers that distinguish AI strings from ordinary strings.
// Require at least 2 matches in a string block to classify as AI.
var aiStringMarkers = regexp.MustCompile(`(?i)\b(you are a|you are an|as an ai|as a helpful|respond with|do not|always respond|your (role|task|job) is|instructions?:|system:|answer the|given the context|based on the|helpful assistant)\b`)

// jsTemplateLiteral finds backtick template literals spanning multiple lines or >60 chars.
var jsTemplateLiteral = regexp.MustCompile("(?s)`([^`]{60,})`")

// jsAssignedString finds const/let/var X = "..." or '...' assignments with long strings.
var jsAssignedString = regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*["']([^"']{60,})["']`)

// jsFewShotArray detects arrays of objects with input/output or user/assistant structure.
var jsFewShotPattern = regexp.MustCompile(`(?s)\[\s*\{\s*(?:input|user|question)\s*:`)

// jsMultiMessageArray detects arrays with multiple role entries.
var jsMultiRolePattern = regexp.MustCompile(`(?s)role\s*:\s*["'](?:system|user|assistant)["']`)

func parseJSPrompts(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}

	add := func(name, reason string, line int, confidence float64, tier string) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          models.SurfaceContext,
			Language:       "js",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: tier,
			Confidence:    confidence,
			Reason:        reason,
		})
	}

	lines := strings.Split(src, "\n")

	// 1. Multi-role message arrays (structural detection).
	// Count role entries — 2+ distinct roles = message array.
	roleMatches := jsMultiRolePattern.FindAllStringIndex(src, -1)
	if len(roleMatches) >= 2 {
		// Find the line of the first match.
		offset := roleMatches[0][0]
		line := 1 + strings.Count(src[:offset], "\n")
		add("message_array", "["+models.DetectorLangChainConstructor+"] structured message array with "+strconv.Itoa(len(roleMatches))+" role entries", line, 0.92, models.TierSemantic)
	}

	// 2. Template literals with AI content.
	for _, m := range jsTemplateLiteral.FindAllStringSubmatchIndex(src, 5) {
		content := src[m[2]:m[3]]
		markerCount := len(aiStringMarkers.FindAllString(content, -1))
		if markerCount >= 2 {
			line := 1 + strings.Count(src[:m[0]], "\n")
			add("template_prompt", "["+models.DetectorContentMarkers+"] template literal with "+strconv.Itoa(markerCount)+" AI instruction markers", line, 0.82, models.TierContent)
			break
		}
	}

	// 3. Assigned string constants with AI content.
	for i, l := range lines {
		if m := jsAssignedString.FindStringSubmatch(l); m != nil {
			varName := m[1]
			strContent := m[2]
			markerCount := len(aiStringMarkers.FindAllString(strContent, -1))
			if markerCount >= 2 {
				add("inline_prompt_"+varName, "["+models.DetectorContentString+"] string constant '"+varName+"' with "+strconv.Itoa(markerCount)+" AI instruction markers", i+1, 0.78, models.TierContent)
				break
			}
		}
	}

	// 4. Few-shot example arrays.
	if loc := jsFewShotPattern.FindStringIndex(src); loc != nil {
		line := 1 + strings.Count(src[:loc[0]], "\n")
		add("few_shot_examples", "["+models.DetectorBracketFewShot+"] array with input/user/question structure (few-shot pattern)", line, 0.85, models.TierSemantic)
	}

	return surfaces
}

// --- Python structured parsing ---

// pyTripleQuote finds triple-quoted strings (""" or ''') with substantial content.
var pyTripleQuote = regexp.MustCompile(`(?s)"""(.{60,}?)"""|'''(.{60,}?)'''`)

// pyFString finds f-strings with AI content.
var pyFString = regexp.MustCompile(`f["']([^"']{60,})["']`)

// pyAssignedString finds variable = "..." assignments.
var pyAssignedString = regexp.MustCompile(`^(\w+)\s*=\s*["']([^"']{60,})["']`)

// pyFewShotPattern detects lists of dicts with input/output structure.
var pyFewShotPattern = regexp.MustCompile(`(?s)\[\s*\{\s*["'](?:input|user|question)["']`)

// pyMultiRolePattern detects dicts with role keys.
var pyMultiRolePattern = regexp.MustCompile(`["']role["']\s*:\s*["'](?:system|user|assistant)["']`)

func parsePythonPrompts(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}

	add := func(name, reason string, line int, confidence float64, tier string) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          models.SurfaceContext,
			Language:       "python",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: tier,
			Confidence:    confidence,
			Reason:        reason,
		})
	}

	// 1. Multi-role message arrays.
	roleMatches := pyMultiRolePattern.FindAllStringIndex(src, -1)
	if len(roleMatches) >= 2 {
		offset := roleMatches[0][0]
		line := 1 + strings.Count(src[:offset], "\n")
		add("message_array", "["+models.DetectorLangChainConstructor+"] structured message array with "+strconv.Itoa(len(roleMatches))+" role entries", line, 0.92, models.TierSemantic)
	}

	// 2. Triple-quote strings with AI content.
	for _, m := range pyTripleQuote.FindAllStringSubmatchIndex(src, 5) {
		var content string
		if m[2] >= 0 && m[3] >= 0 {
			content = src[m[2]:m[3]]
		} else if m[4] >= 0 && m[5] >= 0 {
			content = src[m[4]:m[5]]
		} else {
			continue
		}
		markerCount := len(aiStringMarkers.FindAllString(content, -1))
		if markerCount >= 2 {
			line := 1 + strings.Count(src[:m[0]], "\n")
			add("docstring_prompt", "["+models.DetectorContentMarkers+"] triple-quote string with "+strconv.Itoa(markerCount)+" AI markers", line, 0.80, models.TierContent)
			break
		}
	}

	// 3. Few-shot arrays.
	if loc := pyFewShotPattern.FindStringIndex(src); loc != nil {
		line := 1 + strings.Count(src[:loc[0]], "\n")
		add("few_shot_examples", "["+models.DetectorBracketFewShot+"] list of dicts with input/user/question keys", line, 0.85, models.TierSemantic)
	}

	return surfaces
}

// --- Go structured parsing ---

// goBacktickString finds backtick strings with substantial content.
var goBacktickString = regexp.MustCompile("(?s)`([^`]{60,})`")

// goMultiRolePattern detects map literals or struct literals with role fields.
var goRolePattern = regexp.MustCompile(`(?i)"role"\s*:\s*"(?:system|user|assistant)"`)

func parseGoPrompts(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}

	add := func(name, reason string, line int, confidence float64, tier string) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          models.SurfaceContext,
			Language:       "go",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: tier,
			Confidence:    confidence,
			Reason:        reason,
		})
	}

	// 1. Role-based message structures.
	roleMatches := goRolePattern.FindAllStringIndex(src, -1)
	if len(roleMatches) >= 2 {
		offset := roleMatches[0][0]
		line := 1 + strings.Count(src[:offset], "\n")
		add("message_array", "["+models.DetectorBracketMessageArray+"] Go struct/map with "+strconv.Itoa(len(roleMatches))+" role entries", line, 0.90, models.TierSemantic)
	}

	// 2. Backtick strings with AI content.
	for _, m := range goBacktickString.FindAllStringSubmatchIndex(src, 5) {
		content := src[m[2]:m[3]]
		markerCount := len(aiStringMarkers.FindAllString(content, -1))
		if markerCount >= 2 {
			line := 1 + strings.Count(src[:m[0]], "\n")
			add("backtick_prompt", "["+models.DetectorContentMarkers+"] backtick string with "+strconv.Itoa(markerCount)+" AI markers", line, 0.78, models.TierContent)
			break
		}
	}

	return surfaces
}
