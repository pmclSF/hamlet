package analysis

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileCache provides thread-safe, shared-nothing caching of file contents
// and parsed Go ASTs across analysis stages. This eliminates redundant I/O:
// without caching, a source file can be read 4–5 times (code units, surfaces,
// AI context, RAG, symbols) and a Go file can be AST-parsed 3 times (imports,
// prompts, symbols).
//
// Usage:
//
//	cache := NewFileCache(root)
//	content, ok := cache.ReadFile("src/auth.ts")       // cached after first read
//	goAST, fset, ok := cache.ParseGoFile("src/svc.go") // cached after first parse
//	cache.Stats()                                        // hit/miss counters
type FileCache struct {
	root string

	mu       sync.RWMutex
	contents map[string]fileCacheEntry // relPath → content
	goASTs   map[string]goASTEntry     // relPath → parsed AST

	// Stats for observability.
	hits      int64
	misses    int64
	astHits   int64
	astMisses int64
}

type fileCacheEntry struct {
	content string
	modTime time.Time
	ok      bool // false if file could not be read
}

type goASTEntry struct {
	file *ast.File
	fset *token.FileSet
	ok   bool // false if parsing failed
}

// FileCacheStats holds cache performance counters.
type FileCacheStats struct {
	ContentHits   int64 `json:"contentHits"`
	ContentMisses int64 `json:"contentMisses"`
	ASTHits       int64 `json:"astHits"`
	ASTMisses     int64 `json:"astMisses"`
	CachedFiles   int   `json:"cachedFiles"`
	CachedASTs    int   `json:"cachedASTs"`
}

// NewFileCache creates a file cache rooted at the given directory.
func NewFileCache(root string) *FileCache {
	return &FileCache{
		root:     root,
		contents: make(map[string]fileCacheEntry, 256),
		goASTs:   make(map[string]goASTEntry, 64),
	}
}

// ReadFile returns the content of a file by repository-relative path.
// Results are cached: the second call for the same path returns instantly.
// Returns ("", false) if the file cannot be read.
func (fc *FileCache) ReadFile(relPath string) (string, bool) {
	fc.mu.RLock()
	entry, cached := fc.contents[relPath]
	fc.mu.RUnlock()

	if cached {
		fc.mu.Lock()
		fc.hits++
		fc.mu.Unlock()
		return entry.content, entry.ok
	}

	// Cache miss — read from disk.
	absPath := filepath.Join(fc.root, relPath)
	data, err := os.ReadFile(absPath)

	fc.mu.Lock()
	fc.misses++
	if err != nil {
		fc.contents[relPath] = fileCacheEntry{ok: false}
		fc.mu.Unlock()
		return "", false
	}
	content := string(data)

	// Capture mod time for incremental support.
	var modTime time.Time
	if info, statErr := os.Stat(absPath); statErr == nil {
		modTime = info.ModTime()
	}

	fc.contents[relPath] = fileCacheEntry{
		content: content,
		modTime: modTime,
		ok:      true,
	}
	fc.mu.Unlock()
	return content, true
}

// ParseGoFile returns a parsed Go AST for the given file. Results are cached:
// the second call for the same path returns the same AST without re-parsing.
// Returns (nil, nil, false) if the file cannot be parsed.
func (fc *FileCache) ParseGoFile(relPath string) (*ast.File, *token.FileSet, bool) {
	fc.mu.RLock()
	entry, cached := fc.goASTs[relPath]
	fc.mu.RUnlock()

	if cached {
		fc.mu.Lock()
		fc.astHits++
		fc.mu.Unlock()
		return entry.file, entry.fset, entry.ok
	}

	// Cache miss — parse. First get the content (also cached).
	src, ok := fc.ReadFile(relPath)
	if !ok {
		fc.mu.Lock()
		fc.astMisses++
		fc.goASTs[relPath] = goASTEntry{ok: false}
		fc.mu.Unlock()
		return nil, nil, false
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, relPath, src, parser.ParseComments)

	fc.mu.Lock()
	fc.astMisses++
	if err != nil {
		fc.goASTs[relPath] = goASTEntry{ok: false}
		fc.mu.Unlock()
		return nil, nil, false
	}
	fc.goASTs[relPath] = goASTEntry{file: f, fset: fset, ok: true}
	fc.mu.Unlock()

	return f, fset, true
}

// ParseGoFileImportsOnly returns a Go AST parsed with ImportsOnly mode.
// This is faster than full parsing and sufficient for import graph construction.
// Falls back to the full AST if already cached.
func (fc *FileCache) ParseGoFileImportsOnly(relPath string) (*ast.File, *token.FileSet, bool) {
	// If we already have a full AST cached, use it.
	fc.mu.RLock()
	entry, cached := fc.goASTs[relPath]
	fc.mu.RUnlock()
	if cached {
		fc.mu.Lock()
		fc.astHits++
		fc.mu.Unlock()
		return entry.file, entry.fset, entry.ok
	}

	// Parse with ImportsOnly for speed. Don't cache as full AST since
	// the tree is incomplete — but still avoid duplicate I/O by reading
	// through the content cache.
	src, ok := fc.ReadFile(relPath)
	if !ok {
		return nil, nil, false
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, relPath, src, parser.ImportsOnly)
	if err != nil {
		return nil, nil, false
	}
	return f, fset, true
}

// ModTime returns the cached modification time for a file, or zero if
// the file hasn't been read yet. Used for incremental analysis.
func (fc *FileCache) ModTime(relPath string) time.Time {
	fc.mu.RLock()
	entry, ok := fc.contents[relPath]
	fc.mu.RUnlock()
	if !ok {
		return time.Time{}
	}
	return entry.modTime
}

// IsStale returns true if the file on disk has been modified since it was
// cached. Returns true (conservative) if the file hasn't been cached yet.
func (fc *FileCache) IsStale(relPath string) bool {
	fc.mu.RLock()
	entry, ok := fc.contents[relPath]
	fc.mu.RUnlock()
	if !ok {
		return true
	}

	absPath := filepath.Join(fc.root, relPath)
	info, err := os.Stat(absPath)
	if err != nil {
		return true
	}
	return info.ModTime().After(entry.modTime)
}

// Invalidate removes a specific file from the cache, forcing a re-read
// on the next access. Used for incremental updates.
func (fc *FileCache) Invalidate(relPath string) {
	fc.mu.Lock()
	delete(fc.contents, relPath)
	delete(fc.goASTs, relPath)
	fc.mu.Unlock()
}

// InvalidateStale checks all cached files and removes entries whose
// on-disk modification time is newer than the cached version.
// Returns the list of invalidated paths.
func (fc *FileCache) InvalidateStale() []string {
	fc.mu.RLock()
	paths := make([]string, 0, len(fc.contents))
	for p := range fc.contents {
		paths = append(paths, p)
	}
	fc.mu.RUnlock()

	var stale []string
	for _, p := range paths {
		if fc.IsStale(p) {
			stale = append(stale, p)
		}
	}

	if len(stale) > 0 {
		fc.mu.Lock()
		for _, p := range stale {
			delete(fc.contents, p)
			delete(fc.goASTs, p)
		}
		fc.mu.Unlock()
	}

	return stale
}

// Stats returns cache performance counters.
func (fc *FileCache) Stats() FileCacheStats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return FileCacheStats{
		ContentHits:   fc.hits,
		ContentMisses: fc.misses,
		ASTHits:       fc.astHits,
		ASTMisses:     fc.astMisses,
		CachedFiles:   len(fc.contents),
		CachedASTs:    len(fc.goASTs),
	}
}

// PrewarmSourceFiles reads all source files into the cache in parallel.
// This front-loads I/O so that subsequent analysis stages hit the cache.
func (fc *FileCache) PrewarmSourceFiles(sourceFiles []string) {
	parallelForEachIndex(len(sourceFiles), func(i int) {
		fc.ReadFile(sourceFiles[i])
	})
}

// PrewarmSourceFilesCtx is like PrewarmSourceFiles but respects cancellation.
func (fc *FileCache) PrewarmSourceFilesCtx(ctx context.Context, sourceFiles []string) {
	parallelForEachIndexCtx(ctx, len(sourceFiles), func(i int) {
		fc.ReadFile(sourceFiles[i])
	})
}

// Note: languageForExt, skipDirs, and isJSExt are defined in
// language.go, repository_scan.go, and framework_detection.go respectively.
