package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/slash"
	"github.com/pmclSF/terrain/internal/suppression"
)

// TestRealDispatcher_DismissWritesSuppression is the end-to-end
// integration test for the slash-receiver → CLI runner →
// suppression-file write chain.
//
// The contract: a /dismiss webhook delivery causes a real
// .terrain/suppressions.yaml to land on disk, with the comment
// author recorded as owner. No mocks — runSuppress is the same
// runner the CLI uses.
func TestRealDispatcher_DismissWritesSuppression(t *testing.T) {
	root := t.TempDir()
	d := newRealDispatcher(root)

	ev := slash.WebhookEvent{
		Sender:    "octocat",
		FindingID: "weakAssertion@internal/auth/login_test.go:TestLogin#abc123",
	}
	cmd := &slash.Command{
		Verb:    slash.VerbDismiss,
		Keyword: map[string]string{"reason": "false positive — sanitized upstream"},
	}

	reply, err := d.Handle(ev, cmd)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(reply, "Dismissed") {
		t.Errorf("reply should confirm dismissal; got: %q", reply)
	}

	res, err := suppression.Load(filepath.Join(root, suppression.DefaultPath))
	if err != nil {
		t.Fatalf("load suppressions: %v", err)
	}
	if len(res.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d: %+v", len(res.Entries), res.Entries)
	}
	if !strings.Contains(res.Entries[0].Owner, "octocat") {
		t.Errorf("owner should record comment author; got %q", res.Entries[0].Owner)
	}
}

// TestRealDispatcher_DismissRequiresReason proves the receiver
// enforces the reason: keyword. Reason-less dismisses are a known
// adopter foot-gun (silent suppression with no audit trail).
func TestRealDispatcher_DismissRequiresReason(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	ev := slash.WebhookEvent{FindingID: "x@y:z#1"}
	cmd := &slash.Command{Verb: slash.VerbDismiss}
	if _, err := d.Handle(ev, cmd); err == nil {
		t.Error("expected error when reason: is missing")
	}
}

// TestRealDispatcher_DismissOutsideThreadDeclines covers the
// "comment isn't a reply to a Terrain finding" case. The receiver
// MUST NOT write a suppression without an attached finding-id —
// otherwise the next `analyze` re-emits the same finding and the
// user thinks /dismiss is broken.
func TestRealDispatcher_DismissOutsideThreadDeclines(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	ev := slash.WebhookEvent{Sender: "octocat"} // no FindingID
	cmd := &slash.Command{
		Verb:    slash.VerbDismiss,
		Keyword: map[string]string{"reason": "x"},
	}
	reply, err := d.Handle(ev, cmd)
	if err != nil {
		t.Fatalf("expected graceful decline, got error: %v", err)
	}
	if !strings.Contains(reply, "Cannot dismiss") {
		t.Errorf("reply should explain the decline; got: %q", reply)
	}
}

// TestRealDispatcher_DismissAcceptsFindingKeywordFallback proves the
// user-typed `finding:<id>` keyword bypasses the proxy-injected header.
// This is the manual escape hatch for adopters who haven't deployed
// the X-Terrain-Finding-Id-injecting proxy yet — the conversation
// loop still closes by typing the id directly.
func TestRealDispatcher_DismissAcceptsFindingKeywordFallback(t *testing.T) {
	root := t.TempDir()
	d := newRealDispatcher(root)
	ev := slash.WebhookEvent{Sender: "octocat"} // intentionally no FindingID
	cmd := &slash.Command{
		Verb: slash.VerbDismiss,
		Keyword: map[string]string{
			"reason":  "false positive",
			"finding": "weakAssertion@src/x_test.go:TestX#abc",
		},
	}
	reply, err := d.Handle(ev, cmd)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(reply, "Dismissed weakAssertion@src/x_test.go:TestX#abc") {
		t.Errorf("reply should confirm dismissal of the keyword-supplied id; got: %q", reply)
	}
}

// TestRealDispatcher_UnroutedVerbsGetUnknownCommandReply proves verbs
// the dispatcher does not route fall through to the unknown-command
// reply: no dispatcher error (the webhook must not 500), and the reply
// lists the verbs that do work so the user can self-correct.
func TestRealDispatcher_UnroutedVerbsGetUnknownCommandReply(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	cases := []slash.Verb{
		slash.VerbRefresh,
		slash.VerbEscalate,
		slash.Verb("made-up"),
	}
	for _, v := range cases {
		t.Run(string(v), func(t *testing.T) {
			cmd := &slash.Command{Verb: v}
			reply, err := d.Handle(slash.WebhookEvent{}, cmd)
			if err != nil {
				t.Fatalf("Handle(%s): %v", v, err)
			}
			for _, must := range []string{"Unknown command", "/dismiss", "/terrain commands"} {
				if !strings.Contains(reply, must) {
					t.Errorf("verb %s reply should contain %q; got: %q", v, must, reply)
				}
			}
			if strings.Contains(reply, "future release") {
				t.Errorf("verb %s reply must not promise future behavior; got: %q", v, reply)
			}
		})
	}
}

// TestRealDispatcher_ScaffoldGeneratesFromRealSchema is the end-to-end
// proof that /terrain scaffold accept actually materializes a test
// scaffold from a schema sitting in the repo, not a placeholder reply.
func TestRealDispatcher_ScaffoldGeneratesFromRealSchema(t *testing.T) {
	root := t.TempDir()
	schemaPath := filepath.Join(root, "schemas", "input.json")
	if err := os.MkdirAll(filepath.Dir(schemaPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	schema := []byte(`{"properties": {"query": {"type": "string"}}}`)
	if err := os.WriteFile(schemaPath, schema, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	d := newRealDispatcher(root)
	cmd := &slash.Command{
		Verb:       slash.VerbScaffold,
		Positional: []string{"accept"},
		Keyword:    map[string]string{"schema": "schemas/input.json", "lang": "python"},
	}
	reply, err := d.Handle(slash.WebhookEvent{}, cmd)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(reply, "boundary cases") {
		t.Errorf("reply should mention boundary cases; got: %q", reply)
	}
	if !strings.Contains(reply, "import pytest") {
		t.Errorf("reply should contain generated pytest scaffold; got: %q", reply)
	}
}

// TestRealDispatcher_ScaffoldWithoutSchemaShowsUsage covers the
// missing-keyword path — the reply should tell the user what to pass.
func TestRealDispatcher_ScaffoldWithoutSchemaShowsUsage(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	cmd := &slash.Command{Verb: slash.VerbScaffold, Keyword: map[string]string{}}
	reply, err := d.Handle(slash.WebhookEvent{}, cmd)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(reply, "schema:<path>") {
		t.Errorf("reply should document schema:<path>; got: %q", reply)
	}
}

// TestRealDispatcher_ScaffoldRejectsPathEscape proves the safeJoin
// guard prevents a hostile slash comment from asking the server to
// read files outside the repo root.
func TestRealDispatcher_ScaffoldRejectsPathEscape(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	cmd := &slash.Command{
		Verb:    slash.VerbScaffold,
		Keyword: map[string]string{"schema": "../../etc/passwd"},
	}
	reply, err := d.Handle(slash.WebhookEvent{}, cmd)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(reply, "escapes") {
		t.Errorf("reply should reject path escape; got: %q", reply)
	}
}

// TestRealDispatcher_BenchWithoutIDShowsUsage and with-id documents
// the bench replies are stable: usage without an id, local-run
// guidance (echoing the id) with one.
func TestRealDispatcher_BenchReplies(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	t.Run("without id", func(t *testing.T) {
		cmd := &slash.Command{Verb: slash.VerbBench}
		reply, err := d.Handle(slash.WebhookEvent{}, cmd)
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if !strings.Contains(reply, "requires a benchmark id") {
			t.Errorf("reply should document usage; got: %q", reply)
		}
	})
	t.Run("with id", func(t *testing.T) {
		cmd := &slash.Command{Verb: slash.VerbBench, Positional: []string{"latency-001"}}
		reply, err := d.Handle(slash.WebhookEvent{}, cmd)
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if !strings.Contains(reply, "latency-001") {
			t.Errorf("reply should echo the bench id; got: %q", reply)
		}
	})
}

// TestRealDispatcher_ExplainWithMissingArgErrors covers the
// VerbExplain path's argument validation. Without a positional rule-id
// the dispatcher must return a clear error rather than running a
// full analyze pipeline that ends in "entity not found." This
// guards the captureRun stdout-swap path with a fast-failing case.
func TestRealDispatcher_ExplainWithMissingArgErrors(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	cmd := &slash.Command{Verb: slash.VerbExplain}
	_, err := d.Handle(slash.WebhookEvent{}, cmd)
	if err == nil {
		t.Fatalf("expected error for missing rule-id, got nil")
	}
	if !strings.Contains(err.Error(), "rule-id") {
		t.Errorf("error should mention 'rule-id'; got: %v", err)
	}
}

// TestRealDispatcher_WhyWithMissingArgErrors mirrors ExplainWithMissingArg
// for the VerbWhy alias.
func TestRealDispatcher_WhyWithMissingArgErrors(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	cmd := &slash.Command{Verb: slash.VerbWhy}
	_, err := d.Handle(slash.WebhookEvent{}, cmd)
	if err == nil {
		t.Fatalf("expected error for missing rule-id, got nil")
	}
}

// TestRealDispatcher_ShowWithMissingArgErrors covers VerbShow's
// argument validation path.
func TestRealDispatcher_ShowWithMissingArgErrors(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	cmd := &slash.Command{Verb: slash.VerbShow}
	_, err := d.Handle(slash.WebhookEvent{}, cmd)
	if err == nil {
		t.Fatalf("expected error for missing id, got nil")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("error should mention 'id'; got: %v", err)
	}
}

// TestRealDispatcher_CommandsListIsRendered guards the /terrain
// commands surface. Adopters use this to discover what the bot can
// do. If the list ever drifts to empty / panic, /terrain commands
// would be the canary.
func TestRealDispatcher_CommandsListIsRendered(t *testing.T) {
	d := newRealDispatcher(t.TempDir())
	cmd := &slash.Command{Verb: slash.VerbCommands}
	reply, err := d.Handle(slash.WebhookEvent{}, cmd)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	// Every shipping verb must appear in the rendered list.
	for _, must := range []string{"/dismiss", "/terrain explain", "/terrain show"} {
		if !strings.Contains(reply, must) {
			t.Errorf("commands list missing %q; got:\n%s", must, reply)
		}
	}
}
