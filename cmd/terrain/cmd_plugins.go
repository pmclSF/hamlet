package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pmclSF/terrain/internal/plugin"
	"github.com/pmclSF/terrain/internal/uitokens"
)

// runPlugins dispatches `terrain plugins <verb>`. Two verbs ship today:
//
//	manifest <path>   validate a plugin manifest file
//	list              list registered plugins (today: empty stub)
//
// `add` and `remove` are recognized but return a clear not-available
// error: terrain has no plugin runtime (subprocess spawn, signed-binary
// verification, per-plugin directory layout). The verb dispatch is wired
// so a runtime can attach without a CLI surface change.
func runPlugins(args []string) error {
	if len(args) == 0 {
		printPluginsUsage()
		return fmt.Errorf("missing subcommand")
	}
	verb, rest := args[0], args[1:]
	switch verb {
	case "manifest":
		return runPluginsManifest(rest)
	case "list":
		return runPluginsList(rest)
	case "add", "remove":
		return fmt.Errorf("terrain plugins %s is not available: terrain has "+
			"no plugin runtime (validate a manifest with `terrain plugins manifest`)", verb)
	default:
		printPluginsUsage()
		return fmt.Errorf("unknown verb %q", verb)
	}
}

func printPluginsUsage() {
	fmt.Fprintln(os.Stderr, "Usage: terrain plugins <verb> [args]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Verbs:")
	fmt.Fprintln(os.Stderr, "  manifest <path>   validate a plugin manifest file (YAML)")
	fmt.Fprintln(os.Stderr, "  list              list registered plugins")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Not available (terrain has no plugin runtime):")
	fmt.Fprintln(os.Stderr, "  add <plugin-id>   install + register a plugin")
	fmt.Fprintln(os.Stderr, "  remove <plugin-id>  unregister")
}

func runPluginsManifest(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: terrain plugins manifest <path>")
	}
	path := args[0]
	jsonOut := false
	for _, a := range args[1:] {
		if a == "--json" {
			jsonOut = true
		}
	}
	m, err := plugin.LoadManifest(path)
	if err != nil {
		return fmt.Errorf("validate %s: %w", path, err)
	}
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(m)
	}
	fmt.Printf("%s %s validates against plugin manifest schema v%d.\n", uitokens.GlyphOK(), path, plugin.SchemaVersion)
	fmt.Printf("  ID:          %s\n", m.ID)
	fmt.Printf("  Name:        %s\n", m.Name)
	fmt.Printf("  Version:     %s (author: %s)\n", m.Version, m.Author)
	fmt.Printf("  Detectors:   %d\n", len(m.Detectors))
	for _, d := range m.Detectors {
		tier := d.Tier
		if tier == "" {
			tier = "observability"
		}
		fmt.Printf("    - %s [%s, %s, %s]\n", d.RuleID, d.SignalType, d.MechanismClass, tier)
	}
	if m.RequiresNetwork {
		fmt.Println("  Network:     YES (this plugin makes outbound calls)")
	}
	if len(m.RequiresAPIKey) > 0 {
		fmt.Printf("  API keys:    %v\n", m.RequiresAPIKey)
	}
	return nil
}

func runPluginsList(args []string) error {
	jsonOut := false
	for _, a := range args {
		if a == "--json" {
			jsonOut = true
		}
	}
	// There is no plugin runtime, so the list is always empty — terrain
	// ships zero installed plugins. Returning the shape adopters will see
	// at scale so their tooling doesn't need to change later.
	out := map[string]any{
		"schema_version": plugin.SchemaVersion,
		"plugins":        []any{},
	}
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}
	fmt.Println("No plugins registered.")
	fmt.Println()
	fmt.Println("Validate a third-party plugin manifest with:")
	fmt.Println("  terrain plugins manifest <path-to-plugin-manifest.yaml>")
	fmt.Println()
	fmt.Println("Installing plugins (terrain plugins add) is not available — terrain validates manifests only.")
	return nil
}
