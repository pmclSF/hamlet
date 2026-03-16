package depgraph

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// Build constructs a dependency graph from a TestSuiteSnapshot.
//
// The graph is populated in ten stages:
//  1. Test structure: TestFile → TestCase → Suite hierarchy
//  2. Import edges: TestFile → SourceFile (from ImportGraph)
//  3. Source-to-source edges: SourceFile → SourceFile (from ImportGraph overlap)
//  4. Code surfaces: CodeSurface → SourceFile (from inferred behavior anchors)
//  5. Behavior surfaces: BehaviorSurface → CodeSurface (derived groupings)
//  6. Scenarios: Scenario → CodeSurface/BehaviorSurface (behavioral validation)
//  7. Manual coverage: ManualCoverageArtifact → CodeSurface (overlay validation)
//  8. Environments: Environment nodes from CI config inference
//  9. Environment classes: EnvironmentClass → Environment groupings
//  10. Device configs: DeviceConfig nodes for device/browser targets
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
	buildCodeSurfaces(g, snap)
	buildBehaviorSurfaces(g, snap)
	buildScenarios(g, snap)
	buildManualCoverage(g, snap)
	buildEnvironments(g, snap)
	buildEnvironmentClasses(g, snap)
	buildDeviceConfigs(g, snap)
	buildEnvironmentEdges(g, snap)

	g.Seal() // Enable query caching — graph is now read-only.
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
	// Track existing edges to avoid O(n²) linear scans per pair.
	type edgeKey struct{ from, to string }
	seen := map[edgeKey]bool{}
	for _, e := range g.Edges() {
		if e.Type == EdgeSourceImportsSource {
			seen[edgeKey{e.From, e.To}] = true
		}
	}

	for _, imports := range snap.ImportGraph {
		srcList := make([]string, 0, len(imports))
		for s := range imports {
			srcList = append(srcList, s)
		}

		// Only create source→source edges within the same package to
		// avoid noisy cross-package connections.
		for i := 0; i < len(srcList); i++ {
			for j := i + 1; j < len(srcList); j++ {
				if inferPackage(srcList[i]) == inferPackage(srcList[j]) {
					srcAID := "file:" + srcList[i]
					srcBID := "file:" + srcList[j]

					key := edgeKey{srcAID, srcBID}
					if !seen[key] {
						seen[key] = true
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

// buildCodeSurfaces creates code surface nodes from inferred behavior anchors
// and connects them to their containing source files.
func buildCodeSurfaces(g *Graph, snap *models.TestSuiteSnapshot) {
	if len(snap.CodeSurfaces) == 0 {
		return
	}

	for _, cs := range snap.CodeSurfaces {
		surfaceID := cs.SurfaceID
		if surfaceID == "" {
			continue
		}

		meta := map[string]string{
			"kind":     string(cs.Kind),
			"language": cs.Language,
		}
		if cs.Exported {
			meta["exported"] = "true"
		}
		if cs.HTTPMethod != "" {
			meta["httpMethod"] = cs.HTTPMethod
		}
		if cs.Route != "" {
			meta["route"] = cs.Route
		}
		if cs.Receiver != "" {
			meta["receiver"] = cs.Receiver
		}

		g.AddNode(&Node{
			ID:       surfaceID,
			Type:     NodeCodeSurface,
			Path:     cs.Path,
			Name:     cs.Name,
			Line:     cs.Line,
			Package:  cs.Package,
			Metadata: meta,
		})

		// Connect surface to its containing source file.
		srcFileID := "file:" + cs.Path
		if g.Node(srcFileID) == nil {
			g.AddNode(&Node{
				ID:      srcFileID,
				Type:    NodeSourceFile,
				Path:    cs.Path,
				Name:    filepath.Base(cs.Path),
				Package: inferPackage(cs.Path),
			})
		}
		g.AddEdge(&Edge{
			From:         surfaceID,
			To:           srcFileID,
			Type:         EdgeBelongsToPackage,
			Confidence:   1.0,
			EvidenceType: EvidenceStaticAnalysis,
		})

		// If the surface has a linked code unit, connect it.
		if cs.LinkedCodeUnit != "" {
			g.AddEdge(&Edge{
				From:         surfaceID,
				To:           cs.LinkedCodeUnit,
				Type:         EdgeBehaviorDerivedFrom,
				Confidence:   0.8,
				EvidenceType: EvidenceInferred,
			})
		}
	}
}

// buildBehaviorSurfaces creates behavior surface nodes from derived behavior
// groupings and connects them to their constituent code surfaces.
func buildBehaviorSurfaces(g *Graph, snap *models.TestSuiteSnapshot) {
	if len(snap.BehaviorSurfaces) == 0 {
		return
	}

	for _, bs := range snap.BehaviorSurfaces {
		if bs.BehaviorID == "" {
			continue
		}

		meta := map[string]string{
			"kind": string(bs.Kind),
		}
		if bs.RoutePrefix != "" {
			meta["routePrefix"] = bs.RoutePrefix
		}

		g.AddNode(&Node{
			ID:       bs.BehaviorID,
			Type:     NodeBehaviorSurface,
			Name:     bs.Label,
			Package:  bs.Package,
			Metadata: meta,
		})

		// Connect behavior surface to each constituent code surface.
		for _, csID := range bs.CodeSurfaceIDs {
			g.AddEdge(&Edge{
				From:         bs.BehaviorID,
				To:           csID,
				Type:         EdgeBehaviorDerivedFrom,
				Confidence:   0.7,
				EvidenceType: EvidenceInferred,
			})
		}
	}
}

// buildScenarios creates scenario nodes and connects them to the code
// surfaces and behavior surfaces they validate.
func buildScenarios(g *Graph, snap *models.TestSuiteSnapshot) {
	if len(snap.Scenarios) == 0 {
		return
	}

	for _, sc := range snap.Scenarios {
		if sc.ScenarioID == "" {
			continue
		}

		meta := map[string]string{}
		if sc.Category != "" {
			meta["category"] = sc.Category
		}
		if sc.Framework != "" {
			meta["framework"] = sc.Framework
		}
		if !sc.Executable {
			meta["executable"] = "false"
		}

		g.AddNode(&Node{
			ID:        sc.ScenarioID,
			Type:      NodeScenario,
			Path:      sc.Path,
			Name:      sc.Name,
			Framework: sc.Framework,
			Metadata:  meta,
		})

		// Connect scenario to each surface it covers.
		for _, surfaceID := range sc.CoveredSurfaceIDs {
			g.AddEdge(&Edge{
				From:         sc.ScenarioID,
				To:           surfaceID,
				Type:         EdgeCoversCodeSurface,
				Confidence:   0.8,
				EvidenceType: EvidenceInferred,
			})
		}

		// Connect to owner if specified.
		if sc.Owner != "" {
			ownerID := "owner:" + sc.Owner
			if g.Node(ownerID) == nil {
				g.AddNode(&Node{
					ID:   ownerID,
					Type: NodeOwner,
					Name: sc.Owner,
				})
			}
			g.AddEdge(&Edge{
				From:         ownerID,
				To:           sc.ScenarioID,
				Type:         EdgeOwns,
				Confidence:   1.0,
				EvidenceType: EvidenceManual,
			})
		}
	}
}

// buildManualCoverage creates manual coverage artifact nodes and connects
// them to the surfaces they cover. Manual coverage is an overlay — it
// supplements automated coverage but is never executable CI coverage.
func buildManualCoverage(g *Graph, snap *models.TestSuiteSnapshot) {
	if len(snap.ManualCoverage) == 0 {
		return
	}

	for _, mc := range snap.ManualCoverage {
		if mc.ArtifactID == "" {
			continue
		}

		meta := map[string]string{
			"source": mc.Source,
		}
		if mc.Criticality != "" {
			meta["criticality"] = mc.Criticality
		}
		if mc.Frequency != "" {
			meta["frequency"] = mc.Frequency
		}
		if mc.LastExecuted != "" {
			meta["lastExecuted"] = mc.LastExecuted
		}
		if mc.Area != "" {
			meta["area"] = mc.Area
		}

		g.AddNode(&Node{
			ID:       mc.ArtifactID,
			Type:     NodeManualCoverage,
			Name:     mc.Name,
			Metadata: meta,
		})

		// Connect manual coverage to explicitly listed surfaces.
		for _, surfaceID := range mc.CoveredSurfaceIDs {
			g.AddEdge(&Edge{
				From:         mc.ArtifactID,
				To:           surfaceID,
				Type:         EdgeManualCovers,
				Confidence:   0.7,
				EvidenceType: EvidenceManual,
			})
		}

		// When no explicit surfaces are listed, resolve the area field
		// against packages, services, and behavior surfaces in the graph.
		if len(mc.CoveredSurfaceIDs) == 0 && mc.Area != "" {
			resolveAreaToGraphNodes(g, mc.ArtifactID, mc.Area)
		}

		// Connect to owner if specified.
		if mc.Owner != "" {
			ownerID := "owner:" + mc.Owner
			if g.Node(ownerID) == nil {
				g.AddNode(&Node{
					ID:   ownerID,
					Type: NodeOwner,
					Name: mc.Owner,
				})
			}
			g.AddEdge(&Edge{
				From:         ownerID,
				To:           mc.ArtifactID,
				Type:         EdgeOwns,
				Confidence:   1.0,
				EvidenceType: EvidenceManual,
			})
		}
	}
}

// resolveAreaToGraphNodes matches an area string against behavior surfaces
// and code surfaces in the graph. When a match is found, an
// EdgeManualCovers edge is created with confidence based on match specificity.
func resolveAreaToGraphNodes(g *Graph, artifactID, area string) {
	attachableTypes := []NodeType{NodeBehaviorSurface, NodeCodeSurface}

	// Strip trailing wildcard for prefix matching.
	prefix := area
	isGlob := false
	if len(prefix) > 0 && prefix[len(prefix)-1] == '*' {
		prefix = prefix[:len(prefix)-1]
		isGlob = true
	}

	for _, nodeType := range attachableTypes {
		for _, n := range g.NodesByType(nodeType) {
			if matchesNodeArea(n, prefix) {
				confidence := areaMatchConfidence(n, prefix, isGlob)
				g.AddEdge(&Edge{
					From:         artifactID,
					To:           n.ID,
					Type:         EdgeManualCovers,
					Confidence:   confidence,
					EvidenceType: EvidenceManual,
				})
			}
		}
	}
}

// matchesNodeArea checks if a graph node matches an area prefix.
// It checks the node's Package field, Path field, and Name field.
func matchesNodeArea(n *Node, prefix string) bool {
	if prefix == "" {
		return false
	}
	// Check Package field (most common for area matching).
	if n.Package != "" && strings.HasPrefix(n.Package, prefix) {
		return true
	}
	// Check Path field.
	if n.Path != "" && strings.HasPrefix(n.Path, prefix) {
		return true
	}
	// Check Name field (for services and behavior surfaces).
	if n.Name != "" && strings.HasPrefix(strings.ToLower(n.Name), strings.ToLower(prefix)) {
		return true
	}
	return false
}

// areaMatchConfidence returns the edge confidence based on how specific
// the area match is. More specific matches get higher confidence.
func areaMatchConfidence(n *Node, prefix string, isGlob bool) float64 {
	// Exact match on package or name.
	if n.Package == prefix || strings.EqualFold(n.Name, prefix) {
		return 0.7
	}
	// Path exact match.
	if n.Path == prefix {
		return 0.7
	}
	// Glob/prefix match (less specific).
	if isGlob {
		return 0.5
	}
	// Prefix match (feature-area level).
	return 0.5
}

// buildEnvironmentEdges connects test files and scenarios to the environments
// and devices they target. This runs after all environment, device, test, and
// scenario nodes have been created.
func buildEnvironmentEdges(g *Graph, snap *models.TestSuiteSnapshot) {
	// Connect test files to their target environments and devices.
	for _, tf := range snap.TestFiles {
		fileID := "file:" + tf.Path
		if g.Node(fileID) == nil {
			continue
		}

		for _, envID := range tf.EnvironmentIDs {
			if g.Node(envID) != nil {
				g.AddEdge(&Edge{
					From:         fileID,
					To:           envID,
					Type:         EdgeTargetsEnvironment,
					Confidence:   0.8,
					EvidenceType: EvidenceInferred,
				})
			}
		}

		for _, deviceID := range tf.DeviceIDs {
			if g.Node(deviceID) != nil {
				g.AddEdge(&Edge{
					From:         fileID,
					To:           deviceID,
					Type:         EdgeTargetsEnvironment,
					Confidence:   0.7,
					EvidenceType: EvidenceInferred,
				})
			}
		}
	}

	// Connect scenarios to their target environments.
	for _, sc := range snap.Scenarios {
		if sc.ScenarioID == "" || g.Node(sc.ScenarioID) == nil {
			continue
		}

		for _, envID := range sc.EnvironmentIDs {
			if g.Node(envID) != nil {
				g.AddEdge(&Edge{
					From:         sc.ScenarioID,
					To:           envID,
					Type:         EdgeTargetsEnvironment,
					Confidence:   0.8,
					EvidenceType: EvidenceInferred,
				})
			}
		}
	}
}

// buildEnvironments creates environment nodes from the snapshot's environment
// list. Each environment becomes a NodeEnvironment node. If the environment
// has a ClassID, an EdgeEnvironmentClassContains edge is deferred to
// buildEnvironmentClasses.
func buildEnvironments(g *Graph, snap *models.TestSuiteSnapshot) {
	if len(snap.Environments) == 0 {
		return
	}

	for _, env := range snap.Environments {
		if env.EnvironmentID == "" {
			continue
		}

		meta := map[string]string{}
		if env.OS != "" {
			meta["os"] = env.OS
		}
		if env.OSVersion != "" {
			meta["osVersion"] = env.OSVersion
		}
		if env.Runtime != "" {
			meta["runtime"] = env.Runtime
		}
		if env.CIProvider != "" {
			meta["ciProvider"] = env.CIProvider
		}
		if env.ResourceClass != "" {
			meta["resourceClass"] = env.ResourceClass
		}
		if env.IsProduction {
			meta["isProduction"] = "true"
		}
		if env.InferredFrom != "" {
			meta["inferredFrom"] = env.InferredFrom
		}

		g.AddNode(&Node{
			ID:       env.EnvironmentID,
			Type:     NodeEnvironment,
			Name:     env.Name,
			Metadata: meta,
		})
	}
}

// buildEnvironmentClasses creates environment class nodes and connects them
// to their member environments via EdgeEnvironmentClassContains edges.
func buildEnvironmentClasses(g *Graph, snap *models.TestSuiteSnapshot) {
	if len(snap.EnvironmentClasses) == 0 {
		return
	}

	for _, ec := range snap.EnvironmentClasses {
		if ec.ClassID == "" {
			continue
		}

		meta := map[string]string{}
		if ec.Dimension != "" {
			meta["dimension"] = ec.Dimension
		}

		g.AddNode(&Node{
			ID:       ec.ClassID,
			Type:     NodeEnvironmentClass,
			Name:     ec.Name,
			Metadata: meta,
		})

		// Connect class to each member environment.
		for _, memberID := range ec.MemberIDs {
			if g.Node(memberID) != nil {
				g.AddEdge(&Edge{
					From:         ec.ClassID,
					To:           memberID,
					Type:         EdgeEnvironmentClassContains,
					Confidence:   1.0,
					EvidenceType: EvidenceStaticAnalysis,
				})
			}
		}
	}

	// Also connect environments that declare a ClassID via their own field.
	for _, env := range snap.Environments {
		if env.ClassID != "" && env.EnvironmentID != "" && g.Node(env.ClassID) != nil {
			// Check if edge already exists from the class's MemberIDs.
			alreadyLinked := false
			for _, e := range g.Outgoing(env.ClassID) {
				if e.To == env.EnvironmentID && e.Type == EdgeEnvironmentClassContains {
					alreadyLinked = true
					break
				}
			}
			if !alreadyLinked {
				g.AddEdge(&Edge{
					From:         env.ClassID,
					To:           env.EnvironmentID,
					Type:         EdgeEnvironmentClassContains,
					Confidence:   1.0,
					EvidenceType: EvidenceStaticAnalysis,
				})
			}
		}
	}
}

// buildDeviceConfigs creates device configuration nodes from the snapshot's
// device list. Each device becomes a NodeDeviceConfig node. If the device
// has a ClassID, it is connected to the corresponding EnvironmentClass.
func buildDeviceConfigs(g *Graph, snap *models.TestSuiteSnapshot) {
	if len(snap.DeviceConfigs) == 0 {
		return
	}

	for _, dc := range snap.DeviceConfigs {
		if dc.DeviceID == "" {
			continue
		}

		meta := map[string]string{}
		if dc.Platform != "" {
			meta["platform"] = dc.Platform
		}
		if dc.FormFactor != "" {
			meta["formFactor"] = dc.FormFactor
		}
		if dc.OSVersion != "" {
			meta["osVersion"] = dc.OSVersion
		}
		if dc.BrowserEngine != "" {
			meta["browserEngine"] = dc.BrowserEngine
		}
		if dc.InferredFrom != "" {
			meta["inferredFrom"] = dc.InferredFrom
		}

		g.AddNode(&Node{
			ID:       dc.DeviceID,
			Type:     NodeDeviceConfig,
			Name:     dc.Name,
			Metadata: meta,
		})

		// Connect device to its environment class if specified.
		if dc.ClassID != "" && g.Node(dc.ClassID) != nil {
			g.AddEdge(&Edge{
				From:         dc.ClassID,
				To:           dc.DeviceID,
				Type:         EdgeEnvironmentClassContains,
				Confidence:   1.0,
				EvidenceType: EvidenceStaticAnalysis,
			})
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
	switch parts[0] {
	case "packages", "libs", "apps", "modules":
		return parts[0] + "/" + parts[1]
	}

	// Default: use the first directory.
	return parts[0]
}
