package depgraph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// Build constructs a dependency graph from a TestSuiteSnapshot.
//
// The graph is populated in three stages:
//  1. Test structure: TestFile → TestCase → Suite hierarchy
//  2. Import edges: TestFile → SourceFile (from ImportGraph)
//  3. Source-to-source edges: SourceFile → SourceFile (from ImportGraph overlap)
//
// The resulting graph enables traversal-based analysis that the flat snapshot
// indexes cannot support: coverage via reverse edges, impact via BFS with
// confidence decay, and fanout via transitive closure.
func Build(snap *models.TestSuiteSnapshot) *Graph {
	g := NewGraph()
	if snap == nil {
		return g
	}

	buildTestStructure(g, snap)
	buildImportEdges(g, snap)
	buildSourceToSourceEdges(g, snap)

	return g
}

// buildTestStructure creates test file, test case, and suite nodes with
// their structural edges.
func buildTestStructure(g *Graph, snap *models.TestSuiteSnapshot) {
	for _, tf := range snap.TestFiles {
		fileID := "file:" + tf.Path
		g.AddNode(&Node{
			ID:        fileID,
			Type:      NodeTestFile,
			Path:      tf.Path,
			Name:      filepath.Base(tf.Path),
			Framework: tf.Framework,
			Package:   inferPackage(tf.Path),
		})
	}

	// Group test cases by file for suite hierarchy construction.
	byFile := map[string][]models.TestCase{}
	for _, tc := range snap.TestCases {
		byFile[tc.FilePath] = append(byFile[tc.FilePath], tc)
	}

	for filePath, cases := range byFile {
		fileID := "file:" + filePath
		suitesSeen := map[string]bool{}

		for _, tc := range cases {
			// Create test node.
			testID := fmt.Sprintf("test:%s:%d:%s", tc.FilePath, tc.Line, tc.TestName)
			g.AddNode(&Node{
				ID:        testID,
				Type:      NodeTest,
				Path:      tc.FilePath,
				Name:      tc.TestName,
				Line:      tc.Line,
				Framework: tc.Framework,
				Package:   inferPackage(tc.FilePath),
			})

			// Build suite chain and connect test to file.
			parentID := fileID
			for i, suite := range tc.SuiteHierarchy {
				suiteID := fmt.Sprintf("suite:%s:%s", tc.FilePath, strings.Join(tc.SuiteHierarchy[:i+1], "::"))
				if !suitesSeen[suiteID] {
					suitesSeen[suiteID] = true
					g.AddNode(&Node{
						ID:      suiteID,
						Type:    NodeSuite,
						Path:    tc.FilePath,
						Name:    suite,
						Package: inferPackage(tc.FilePath),
					})
					// Connect suite to parent (file or outer suite).
					g.AddEdge(&Edge{
						From:         parentID,
						To:           suiteID,
						Type:         EdgeSuiteContainsTest,
						Confidence:   1.0,
						EvidenceType: EvidenceStaticAnalysis,
					})
				}
				parentID = suiteID
			}

			// Connect test to its parent (innermost suite or file).
			g.AddEdge(&Edge{
				From:         testID,
				To:           fileID,
				Type:         EdgeTestDefinedInFile,
				Confidence:   1.0,
				EvidenceType: EvidenceStaticAnalysis,
			})
		}
	}
}

// buildImportEdges creates source file nodes and test→source import edges
// from the snapshot's ImportGraph.
func buildImportEdges(g *Graph, snap *models.TestSuiteSnapshot) {
	if snap.ImportGraph == nil {
		return
	}

	// Also create nodes for code units if they exist.
	codeUnitPaths := map[string]bool{}
	for _, cu := range snap.CodeUnits {
		codeUnitPaths[cu.Path] = true
	}

	for testPath, imports := range snap.ImportGraph {
		fileID := "file:" + testPath

		// Ensure test file node exists (it may not if the file wasn't in TestFiles).
		if g.Node(fileID) == nil {
			g.AddNode(&Node{
				ID:      fileID,
				Type:    NodeTestFile,
				Path:    testPath,
				Name:    filepath.Base(testPath),
				Package: inferPackage(testPath),
			})
		}

		for srcPath := range imports {
			srcID := "file:" + srcPath

			// Create source file node if it doesn't exist.
			if g.Node(srcID) == nil {
				g.AddNode(&Node{
					ID:      srcID,
					Type:    NodeSourceFile,
					Path:    srcPath,
					Name:    filepath.Base(srcPath),
					Package: inferPackage(srcPath),
				})
			}

			// Test file imports source file.
			g.AddEdge(&Edge{
				From:         fileID,
				To:           srcID,
				Type:         EdgeImportsModule,
				Confidence:   1.0,
				EvidenceType: EvidenceStaticAnalysis,
			})
		}
	}
}

// buildSourceToSourceEdges infers source-to-source import relationships.
//
// When multiple test files import the same source, and those test files also
// import other shared sources, we infer structural relationships between
// source modules. This is a heuristic — true source-to-source imports
// would require parsing the source files themselves.
func buildSourceToSourceEdges(g *Graph, snap *models.TestSuiteSnapshot) {
	if snap.ImportGraph == nil {
		return
	}

	// Build reverse index: source → set of test files that import it.
	srcToTests := map[string]map[string]bool{}
	for testPath, imports := range snap.ImportGraph {
		for srcPath := range imports {
			if srcToTests[srcPath] == nil {
				srcToTests[srcPath] = map[string]bool{}
			}
			srcToTests[srcPath][testPath] = true
		}
	}

	// For each pair of sources imported by the same test file, create an
	// inferred edge if they share enough test importers.
	// This is kept lightweight — only considers co-imports from same test.
	for testPath, imports := range snap.ImportGraph {
		srcList := make([]string, 0, len(imports))
		for s := range imports {
			srcList = append(srcList, s)
		}
		_ = testPath // used to scope the loop

		// Only create source→source edges within the same package to
		// avoid noisy cross-package connections.
		for i := 0; i < len(srcList); i++ {
			for j := i + 1; j < len(srcList); j++ {
				if inferPackage(srcList[i]) == inferPackage(srcList[j]) {
					srcAID := "file:" + srcList[i]
					srcBID := "file:" + srcList[j]

					// Check if this edge already exists.
					exists := false
					for _, e := range g.Outgoing(srcAID) {
						if e.To == srcBID && e.Type == EdgeSourceImportsSource {
							exists = true
							break
						}
					}
					if !exists {
						g.AddEdge(&Edge{
							From:         srcAID,
							To:           srcBID,
							Type:         EdgeSourceImportsSource,
							Confidence:   0.5,
							EvidenceType: EvidenceInferred,
						})
					}
				}
			}
		}
	}
}

// inferPackage extracts a package identifier from a file path.
// For JS/TS this is typically the first directory; for monorepos it
// includes the package name (e.g., "packages/compiler-core").
func inferPackage(filePath string) string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	if len(parts) <= 1 {
		return ""
	}

	// Handle monorepo patterns: packages/X, libs/X, apps/X.
	if len(parts) >= 2 {
		switch parts[0] {
		case "packages", "libs", "apps", "modules":
			if len(parts) >= 2 {
				return parts[0] + "/" + parts[1]
			}
		}
	}

	// Default: use the first directory.
	return parts[0]
}
