package promptcontract

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/pmclSF/terrain/internal/gitignore"
)

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "venv": true, ".venv": true, "env": true,
	"dist": true, "build": true, "__pycache__": true, "site-packages": true,
	".tox": true, ".terrain": true, "vendor": true,
}

// maxSourceFileSize caps the bytes read and parsed per file (256 KB). Real
// prompt/schema code is a few KB; a larger .py is generated or vendored, not
// hand-written. The cap bounds two costs: the read buffer (a symlink to
// /dev/zero or a huge file can't grow memory unbounded) and the tree-sitter
// parse, whose cost is superlinear in file size — an uncapped multi-MB file
// would freeze the first-run report for minutes.
const maxSourceFileSize = 256 * 1024

// Inventory counts the AI surfaces the analyzer parsed from a repo: the files
// containing prompt surfaces, the prompt surfaces themselves, and the in-repo
// schema definitions.
type Inventory struct {
	PromptFiles int // distinct files containing at least one prompt surface
	Prompts     int // prompt surfaces (interpolated strings / templates)
	Schemas     int // in-repo schema definitions (pydantic / dataclass)
}

// AnalyzeInRepo walks a repository, extracts schemas and prompt surfaces from
// every source file, and returns the schema↔prompt drift — WITHOUT requiring a
// git base ref (diff-free static consistency), so it fires on a plain analyze.
// The AI-context gate is repo-scoped: if no file imports an AI SDK, the repo is
// not analyzed (keeps non-AI code silent while still binding across files, since
// a prompt and its schema often live in different modules).
func AnalyzeInRepo(root string) ([]Drift, error) {
	_, drift, err := AnalyzeRepo(root)
	return drift, err
}

// AnalyzeRepo is AnalyzeInRepo plus the surface inventory it parsed, for
// callers (the first-run report) that show what Terrain understood alongside
// the drift it found.
func AnalyzeRepo(root string) (Inventory, []Drift, error) {
	type pySource struct {
		rel string
		src []byte
	}
	var files []pySource
	maybeAI := false // cheap substring pre-filter — never under-matches a real import

	// Honour the repo .gitignore, matching the analysis walker. A prompt or
	// schema inside a gitignored tree is not the repo's own surface, and any
	// drift there is not the adopter's to fix.
	ign := gitignore.Load(root)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries; never fail the whole walk
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if ign.Match(relOrPath(root, path), true) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".py") {
			return nil
		}
		if ign.Match(relOrPath(root, path), false) {
			return nil
		}
		// Reject anything that isn't a regular file within the cap before
		// reading. WalkDir reports an entry's own type without following
		// symlinks, so a symlink to /dev/zero, a device, or an oversize file is
		// rejected here — os.ReadFile would otherwise follow the link and grow
		// the buffer unbounded, and a huge real file would stall the parse.
		info, infoErr := d.Info()
		if infoErr != nil || !info.Mode().IsRegular() || info.Size() > maxSourceFileSize {
			return nil
		}
		src, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		files = append(files, pySource{relOrPath(root, path), src})
		if !maybeAI && looksLikeAISource(src) {
			maybeAI = true
		}
		return nil
	})
	if err != nil {
		return Inventory{}, nil, err
	}
	// Fast negative gate: a real `import openai` always contains the root as a
	// substring, so if NO file mentions any AI root we can skip tree-sitter
	// parsing entirely. Non-AI repos (the common case) pay only the file read.
	if !maybeAI {
		return Inventory{}, nil, nil
	}

	// Parse files concurrently — tree-sitter parsing dominates the cost and
	// the parser pool (sync.Pool) is concurrency-safe. Results are collected by
	// index so the merge order stays deterministic regardless of scheduling.
	parsed := make([]pyFile, len(files))
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			// A panic inside the CGO tree-sitter parse must not crash the run —
			// the file simply contributes no surfaces (parsed[i] stays zero).
			defer func() { _ = recover() }()
			parsed[i] = extractPython(files[i].rel, files[i].src)
		}(i)
	}
	wg.Wait()

	var schemas []SchemaDef
	var prompts []PromptSurface
	anyAI := false // authoritative per-file gate (parses to confirm a real import)
	for _, pf := range parsed {
		schemas = append(schemas, pf.schemas...)
		prompts = append(prompts, pf.prompts...)
		anyAI = anyAI || pf.hasAI
	}
	if !anyAI {
		return Inventory{}, nil, nil // AI gate: no confirmed AI import -> no drift
	}
	promptFiles := map[string]bool{}
	for _, p := range prompts {
		promptFiles[p.Path] = true
	}
	inv := Inventory{
		PromptFiles: len(promptFiles),
		Prompts:     len(prompts),
		Schemas:     len(schemas),
	}
	return inv, Detect(schemas, prompts), nil
}

// looksLikeAISource is a cheap, over-inclusive substring check for an AI import
// root — a fast negative pre-filter only. It may return true for a file that
// merely mentions a root in a comment or string (harmless: that repo just gets
// parsed and then held to the authoritative pyHasAIImport gate). It never
// returns false for a file that actually imports an AI SDK, because the import
// statement text contains the root verbatim.
func looksLikeAISource(src []byte) bool {
	if !bytes.Contains(src, []byte("import")) {
		return false
	}
	for root := range aiImportRoots {
		if bytes.Contains(src, []byte(root)) {
			return true
		}
	}
	return false
}

func relOrPath(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(path)
}
