package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/findinghistory"
	"github.com/pmclSF/terrain/internal/findings"
	"github.com/pmclSF/terrain/internal/identity"
	"github.com/pmclSF/terrain/internal/scaffold"
	"github.com/pmclSF/terrain/internal/slash"
)

// realDispatcher implements slash.Dispatcher by routing each verb to
// the existing CLI runner that already implements the behavior. This
// is what production webhook deployments use; DefaultDispatcher
// (informational only) stays in internal/slash so the package can be
// tested without depending on cmd/terrain.
//
// Routing:
//
//	/dismiss              → runSuppress (writes a content-hash entry)
//	/terrain show <id>    → runReportShow (renders one finding)
//	/terrain explain <id> → runReportExplain (long-form rule docs)
//	/terrain why <id>     → runReportExplain --short
//	/terrain commands     → slash.RenderCommandList
//	/terrain expand       → renders the full findings list from the
//	                         last analyze artifact
//	/terrain scaffold     → scaffold.GenerateFromSchema + scaffold.Emit
//	/terrain bench        → points at the local `terrain ai run` flow
//
// Verbs the dispatcher does not route get the unknown-command reply,
// which lists the verbs that do work.
//
// The repo root that runners need defaults to ".". A future deployment
// that routes multiple PRs through one server will compute the root
// per-event from the GitHub clone path.
type realDispatcher struct {
	repoRoot string
}

func newRealDispatcher(repoRoot string) *realDispatcher {
	if repoRoot == "" {
		repoRoot = "."
	}
	return &realDispatcher{repoRoot: repoRoot}
}

// Handle implements slash.Dispatcher.
//
// Contract:
//
//	D1. dismiss / show / explain / why / scaffold / commands / expand
//	    execute their real behavior (not placeholders).
//	D2. Verbs with no PR-side behavior (bench) and unrouted verbs get
//	    a polite reply, never a dispatcher error (the webhook must not
//	    500 on them).
//	D3. User-input errors surface as a polite markdown reply, not an error.
func (d *realDispatcher) Handle(ev slash.WebhookEvent, cmd *slash.Command) (string, error) {
	if cmd == nil {
		return "", fmt.Errorf("nil command")
	}
	switch cmd.Verb {
	case slash.VerbCommands:
		return slash.RenderCommandList(), nil

	case slash.VerbDismiss:
		reason := cmd.Keyword["reason"]
		if reason == "" {
			return "", fmt.Errorf("/dismiss requires reason:<text>")
		}
		// Finding-id resolution order:
		//   1. The X-Terrain-Finding-Id header (proxy-injected from
		//      the parent comment's hidden marker).
		//   2. `finding:<id>` keyword on the slash command itself
		//      (user-typed fallback when the proxy is not configured).
		findingID := ev.FindingID
		if findingID == "" {
			findingID = cmd.Keyword["finding"]
		}
		if findingID == "" {
			return "Cannot dismiss: this command must be a reply to a Terrain finding's inline comment, or include `finding:<id>` (copy the ID from the finding card).", nil
		}
		// Default scope=instance with auto content-hash; default
		// expiry per scope. Owner is the comment author.
		owner := ""
		if ev.Sender != "" {
			owner = "@" + ev.Sender
		}
		if err := runSuppress(findingID, reason, "", owner, "instance", d.repoRoot); err != nil {
			// User-input errors (malformed finding-id, missing reason)
			// must NOT bubble up as a dispatcher error — that turns
			// into a 500 from the webhook handler, and GitHub's
			// delivery retry storms a misclick into many duplicate
			// retries. Surface the cause as a polite markdown reply
			// so the user can fix and re-post.
			if isUserInputError(err) {
				return fmt.Sprintf("Cannot dismiss: %s", err.Error()), nil
			}
			return "", fmt.Errorf("runSuppress: %w", err)
		}
		// Reset the per-repo finding-history demote state so the
		// next fire of this (rule, file) doesn't render with a stale
		// "demoted to footer" badge. Failure here is non-fatal —
		// the suppression has already been written and the user
		// gets the visible dismissal.
		if err := recordDismissInHistory(findingID, d.repoRoot); err != nil {
			// Log to stderr only; don't surface as a slash reply
			// since the suppress already succeeded.
			fmt.Fprintf(os.Stderr, "warn: finding-history dismiss failed: %v\n", err)
		}
		return fmt.Sprintf("Dismissed %s (scope=instance). Reason: %q.", findingID, reason), nil

	case slash.VerbExplain, slash.VerbWhy:
		if len(cmd.Positional) == 0 {
			return "", fmt.Errorf("/%s requires a rule-id", cmd.Verb)
		}
		ruleID := strings.Join(cmd.Positional, " ")
		verbose := cmd.Verb == slash.VerbExplain // /terrain why is the short form
		return d.captureRun(func() error {
			return runExplain(ruleID, d.repoRoot, "", false, verbose)
		})

	case slash.VerbShow:
		if len(cmd.Positional) == 0 {
			return "", fmt.Errorf("/terrain show requires an id")
		}
		id := strings.Join(cmd.Positional, " ")
		return d.captureRun(func() error {
			return runShow("finding", id, d.repoRoot, false)
		})

	case slash.VerbExpand:
		return d.handleExpand()

	case slash.VerbScaffold:
		return d.handleScaffold(cmd)

	case slash.VerbBench:
		id := strings.Join(cmd.Positional, " ")
		if id == "" {
			return "_/terrain bench requires a benchmark id (usage: /terrain bench <id>)._", nil
		}
		return fmt.Sprintf("_Benchmarks don't run from PR comments — run %s locally with `terrain ai run <scenario>`._", id), nil
	}
	return fmt.Sprintf("_Unknown command `%s`. Available: `/dismiss reason:<text>`, `/terrain show <id>`, `/terrain explain <rule-id>`, `/terrain why <rule-id>`, `/terrain expand`, `/terrain scaffold accept`, `/terrain commands`._", cmd.Verb), nil
}

// captureRun redirects the runner's stdout to a buffer for the
// duration of fn and returns the captured output as the slash reply.
// Runners like runExplain print directly to stdout; the webhook
// reply needs that text as markdown.
//
// Not safe for concurrent dispatch across goroutines (the os.Stdout
// swap is process-global). The webhook handler serializes incoming
// requests so this is fine in practice.
func (d *realDispatcher) captureRun(fn func() error) (string, error) {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	runErr := fn()
	_ = w.Close()
	<-done
	os.Stdout = orig
	if runErr != nil {
		return "", runErr
	}
	out := strings.TrimSpace(buf.String())
	if out == "" {
		out = "_(no output)_"
	}
	// Wrap the captured text in a markdown code fence so terminal
	// formatting (rules, badges) renders predictably on GitHub.
	return "```\n" + out + "\n```", nil
}

// resolveRepoRoot picks the repo root for a webhook event. For now,
// defaults to the process's cwd. Future multi-repo deployments will
// clone the PR's commit to a temp directory per-event and pass that
// path here.
func resolveRepoRoot() string {
	if cwd, err := filepath.Abs("."); err == nil {
		return cwd
	}
	return "."
}

// isUserInputError reports whether err looks like a malformed-input
// failure that the user can fix in a follow-up comment, rather than
// a transient or internal failure that should retry. The classifier
// is intentionally pattern-based on the suppression package's known
// error messages — there's no error sentinel to match against.
func isUserInputError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, marker := range []string{
		"invalid finding ID format",
		"missing finding ID",
		"--reason is required",
		"invalid --scope",
		"does not look like YYYY-MM-DD",
	} {
		if strings.Contains(msg, marker) {
			return true
		}
	}
	return false
}

// handleExpand renders the full findings list behind a collapsed
// `+N more` block by reading the canonical .terrain/findings.json
// artifact written by the last analyze run. Missing/empty artifact is
// reported politely rather than as a dispatcher error.
func (d *realDispatcher) handleExpand() (string, error) {
	path := filepath.Join(d.repoRoot, ".terrain", "findings.json")
	f, err := os.Open(path)
	if err != nil {
		return "_No `.terrain/findings.json` yet — run `terrain analyze` first, then `/terrain expand`._", nil
	}
	defer func() { _ = f.Close() }()
	art, err := findings.ReadArtifact(f)
	if err != nil {
		return fmt.Sprintf("_Could not read findings.json: %s._", err.Error()), nil
	}
	if art == nil || len(art.Findings) == 0 {
		return "_No findings — nothing to expand._", nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "**All %d findings** (the full list behind `+N more`):\n\n", len(art.Findings))
	for _, fnd := range art.Findings {
		loc := fnd.PrimaryLoc.Path
		if fnd.PrimaryLoc.Line > 0 {
			loc = fmt.Sprintf("%s:%d", loc, fnd.PrimaryLoc.Line)
		}
		fmt.Fprintf(&b, "- **[%s] %s** `%s` — %s\n", fnd.Severity, fnd.RuleID, loc, fnd.ShortMessage)
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

// handleScaffold runs the mutation-test scaffold generator from a
// slash invocation. Keyword args:
//
//	schema:<path>   required — JSON Schema relative to repo root
//	prompt:<path>   optional — prompt under test (header comment)
//	lang:python|typescript|json   optional — default python
//
// The schema path is resolved under repoRoot and rejected if it
// escapes (so a hostile slash comment can't ask the server to read
// arbitrary files). Returns the rendered scaffold inside a markdown
// fenced block; adopters copy-paste into their tests/ tree.
func (d *realDispatcher) handleScaffold(cmd *slash.Command) (string, error) {
	schemaArg := cmd.Keyword["schema"]
	if schemaArg == "" {
		return "_/terrain scaffold accept needs `schema:<path>`. Example: `/terrain scaffold accept schema:schemas/input.json prompt:prompts/main.md lang:python`._", nil
	}
	resolved, err := safeJoin(d.repoRoot, schemaArg)
	if err != nil {
		return fmt.Sprintf("_Cannot read `%s`: %s._", schemaArg, err.Error()), nil
	}
	body, err := os.ReadFile(resolved)
	if err != nil {
		return fmt.Sprintf("_Cannot read schema `%s`: %s._", schemaArg, err.Error()), nil
	}
	cases, err := scaffold.GenerateFromSchema(body)
	if err != nil {
		return fmt.Sprintf("_Schema `%s` did not parse: %s._", schemaArg, err.Error()), nil
	}
	if len(cases) == 0 {
		return fmt.Sprintf("_Schema `%s` declared no `properties` — nothing to scaffold._", schemaArg), nil
	}
	lang := cmd.Keyword["lang"]
	if lang == "" {
		lang = "python"
	}
	out := scaffold.Emit(cases, scaffold.EmitOptions{
		SchemaPath: schemaArg,
		PromptPath: cmd.Keyword["prompt"],
		Language:   lang,
	})
	fence := "```python"
	switch lang {
	case "typescript", "ts":
		fence = "```typescript"
	case "json":
		fence = "```json"
	}
	return fmt.Sprintf("Generated mutation-test scaffold for `%s` (%d boundary cases). Drop into your tests/ tree:\n\n%s\n%s\n```", schemaArg, len(cases), fence, strings.TrimRight(out, "\n")), nil
}

// safeJoin resolves a relative path under root, rejecting anything
// that escapes root via "..". Used to prevent webhook callers from
// asking the server to read arbitrary files.
func safeJoin(root, rel string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute paths not allowed")
	}
	resolved := filepath.Join(root, rel)
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolvedAbs, err := filepath.Abs(resolved)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(resolvedAbs+string(filepath.Separator), rootAbs+string(filepath.Separator)) && resolvedAbs != rootAbs {
		return "", fmt.Errorf("path escapes repo root")
	}
	// A lexical check alone is not enough: a symlink inside the repo can
	// point outside it, so re-verify containment against the fully
	// symlink-resolved path. When the target does not exist yet,
	// EvalSymlinks errors — the lexical check above already rejected ".."
	// escapes, and the later ReadFile surfaces the not-found error.
	realRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		realRoot = rootAbs
	}
	realResolved, err := filepath.EvalSymlinks(resolvedAbs)
	if err == nil {
		if !strings.HasPrefix(realResolved+string(filepath.Separator), realRoot+string(filepath.Separator)) && realResolved != realRoot {
			return "", fmt.Errorf("path escapes repo root")
		}
	}
	return resolvedAbs, nil
}

// recordDismissInHistory writes the user's dismissal to the per-repo
// finding-history store so the renderer's demote machinery treats this
// as an active user signal (LastDismiss). Without this, the
// dismiss-overrides-fatigue-demotion contract — documented on
// findinghistory.Store.ShouldDemote — never fires in production,
// because nothing else in the CLI writes LastDismiss.
//
// Returns nil when the finding-id can't be parsed (the suppression
// path uses a different format, e.g. file-scope) or when the history
// store can't be loaded; in both cases the suppression still happened.
func recordDismissInHistory(findingID, repoRoot string) error {
	ruleID, filePath, _, _, ok := identity.ParseFindingID(findingID)
	if !ok || ruleID == "" || filePath == "" {
		return nil
	}
	store, err := engine.LoadFindingHistory(repoRoot)
	if err != nil {
		return err
	}
	store.Dismiss(ruleID, filePath)
	return store.Save(filepath.Join(repoRoot, findinghistory.DefaultPath))
}
