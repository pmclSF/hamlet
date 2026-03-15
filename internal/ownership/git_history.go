package ownership

import (
	"bufio"
	"bytes"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const gitHistoryAuthorPrefix = "__TERRAIN_AUTHOR__:"

func (r *Resolver) matchGitHistoryAssignment(relPath string) (OwnershipAssignment, bool) {
	if !r.shouldTryGitHistory() {
		return OwnershipAssignment{}, false
	}

	if !r.gitHistoryLoaded {
		r.loadGitHistoryOwners()
	}

	normalized := filepath.ToSlash(relPath)
	ownerID, ok := r.gitHistoryOwners[normalized]
	if !ok || ownerID == "" {
		return OwnershipAssignment{}, false
	}

	return OwnershipAssignment{
		Owners:      []Owner{{ID: ownerID}},
		Source:      SourceGitHistory,
		Confidence:  ConfidenceLow,
		Inheritance: InheritanceDirect,
		MatchedRule: "git-log-latest-author",
		SourceFile:  "git-history",
	}, true
}

func (r *Resolver) loadGitHistoryOwners() {
	r.gitHistoryLoaded = true
	r.gitHistoryOwners = map[string]string{}

	maxLogs := r.gitHistoryMaxLogs
	if maxLogs <= 0 {
		maxLogs = 1000
	}

	cmd := exec.Command(
		"git",
		"-C", r.repoRoot,
		"log",
		"--no-merges",
		"--format="+gitHistoryAuthorPrefix+"%ae",
		"--name-only",
		"-n", strconv.Itoa(maxLogs),
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := "failed to resolve git-history owners: " + err.Error()
		if text := strings.TrimSpace(stderr.String()); text != "" {
			msg += ": " + text
		}
		r.diagnostics = append(r.diagnostics, Diagnostic{
			Level:   "warning",
			Message: msg,
			Source:  "git-history",
		})
		return
	}

	scanner := bufio.NewScanner(&stdout)
	currentAuthor := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, gitHistoryAuthorPrefix) {
			currentAuthor = normalizeGitOwnerID(strings.TrimSpace(strings.TrimPrefix(line, gitHistoryAuthorPrefix)))
			continue
		}

		if currentAuthor == "" {
			continue
		}

		relPath := filepath.ToSlash(line)
		if _, exists := r.gitHistoryOwners[relPath]; exists {
			continue
		}
		r.gitHistoryOwners[relPath] = currentAuthor
	}

	if err := scanner.Err(); err != nil {
		r.diagnostics = append(r.diagnostics, Diagnostic{
			Level:   "warning",
			Message: "failed to parse git history ownership fallback: " + err.Error(),
			Source:  "git-history",
		})
	}
}

func normalizeGitOwnerID(raw string) string {
	owner := strings.TrimSpace(raw)
	if owner == "" {
		return ""
	}

	// Prefer username-like IDs: normalize common email formats to a stable handle.
	if at := strings.Index(owner, "@"); at > 0 {
		local := owner[:at]
		domain := strings.ToLower(owner[at+1:])

		switch {
		case domain == "users.noreply.github.com":
			// GitHub noreply emails are either "username@..." or "12345+username@...".
			// Prefer the visible username segment.
			if plus := strings.Index(local, "+"); plus >= 0 && plus+1 < len(local) {
				local = local[plus+1:]
			}
		default:
			// For regular emails, drop plus aliases ("alice+ci@company.com" -> "alice").
			if plus := strings.Index(local, "+"); plus > 0 {
				local = local[:plus]
			}
		}

		owner = local
	}

	owner = strings.TrimSpace(strings.ToLower(owner))
	owner = strings.ReplaceAll(owner, " ", "-")
	return NormalizeOwnerID(owner)
}
