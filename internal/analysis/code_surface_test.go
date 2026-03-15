package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- JS/TS Surface Inference ---

func TestJSSurfaceExtractor_Routes(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"js-app")

	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/routes/api.ts")

	byID := indexSurfaces(surfaces)

	// Should detect route registrations.
	routes := filterByKind(surfaces, models.SurfaceRoute)
	if len(routes) < 3 {
		t.Fatalf("expected at least 3 routes, got %d: %v", len(routes), surfaceNames(surfaces))
	}

	// Verify specific routes.
	assertSurfaceExists(t, byID, "surface:src/routes/api.ts:GET /api/users", models.SurfaceRoute)
	assertSurfaceExists(t, byID, "surface:src/routes/api.ts:POST /api/login", models.SurfaceRoute)
	assertSurfaceExists(t, byID, "surface:src/routes/api.ts:DELETE /api/users/:id", models.SurfaceRoute)

	// Verify route metadata.
	getUsersRoute := byID["surface:src/routes/api.ts:GET /api/users"]
	if getUsersRoute.HTTPMethod != "GET" {
		t.Errorf("expected GET method, got %q", getUsersRoute.HTTPMethod)
	}
	if getUsersRoute.Route != "/api/users" {
		t.Errorf("expected /api/users route, got %q", getUsersRoute.Route)
	}
}

func TestJSSurfaceExtractor_HandlersAndFunctions(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"js-app")

	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/handlers/auth.ts")

	byID := indexSurfaces(surfaces)

	// loginHandler should be detected as a handler (name contains Handler).
	assertSurfaceExists(t, byID, "surface:src/handlers/auth.ts:loginHandler", models.SurfaceHandler)

	// authMiddleware should be detected as a handler (name contains Middleware).
	assertSurfaceExists(t, byID, "surface:src/handlers/auth.ts:authMiddleware", models.SurfaceHandler)

	// Verify linked code unit.
	handler := byID["surface:src/handlers/auth.ts:loginHandler"]
	if handler.LinkedCodeUnit != "src/handlers/auth.ts:loginHandler" {
		t.Errorf("expected linked code unit, got %q", handler.LinkedCodeUnit)
	}
}

func TestJSSurfaceExtractor_ExportedFunctionsAndClasses(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"js-app")

	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/utils/validation.ts")

	byID := indexSurfaces(surfaces)

	// Exported functions.
	assertSurfaceExists(t, byID, "surface:src/utils/validation.ts:validateEmail", models.SurfaceFunction)
	assertSurfaceExists(t, byID, "surface:src/utils/validation.ts:sanitizeInput", models.SurfaceFunction)

	// Exported class.
	assertSurfaceExists(t, byID, "surface:src/utils/validation.ts:Validator", models.SurfaceClass)

	// All should be marked exported.
	for _, s := range surfaces {
		if !s.Exported {
			t.Errorf("surface %q should be exported", s.SurfaceID)
		}
	}
}

// --- Go Surface Inference ---

func TestGoSurfaceExtractor_HandlersAndMethods(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"go-app")

	ext := &goSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "handlers/auth.go")

	byID := indexSurfaces(surfaces)

	// LoginHandler detected as handler (name suffix).
	assertSurfaceExists(t, byID, "surface:handlers/auth.go:LoginHandler", models.SurfaceHandler)

	// UserController methods with http.ResponseWriter detected as handlers.
	assertSurfaceExists(t, byID, "surface:handlers/auth.go:UserController.GetUser", models.SurfaceHandler)
	assertSurfaceExists(t, byID, "surface:handlers/auth.go:UserController.DeleteUser", models.SurfaceHandler)

	// internalHelper is unexported — should NOT appear.
	for _, s := range surfaces {
		if s.Name == "internalHelper" {
			t.Error("unexported function internalHelper should not be a surface")
		}
	}
}

func TestGoSurfaceExtractor_ExportedFunctions(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"go-app")

	ext := &goSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "services/user.go")

	byID := indexSurfaces(surfaces)

	// Exported function.
	assertSurfaceExists(t, byID, "surface:services/user.go:NewUserService", models.SurfaceFunction)

	// Exported methods.
	assertSurfaceExists(t, byID, "surface:services/user.go:UserService.FindByID", models.SurfaceMethod)
	assertSurfaceExists(t, byID, "surface:services/user.go:UserService.Create", models.SurfaceMethod)
}

// --- Python Surface Inference ---

func TestPythonSurfaceExtractor_Routes(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"python-app")

	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "views/auth.py")

	byID := indexSurfaces(surfaces)

	// Route decorators.
	assertSurfaceExists(t, byID, "surface:views/auth.py:POST /api/login", models.SurfaceRoute)
	assertSurfaceExists(t, byID, "surface:views/auth.py:GET /api/users", models.SurfaceRoute)

	// Public functions.
	assertSurfaceExists(t, byID, "surface:views/auth.py:login_handler", models.SurfaceHandler)
	assertSurfaceExists(t, byID, "surface:views/auth.py:validate_token", models.SurfaceFunction)
}

func TestPythonSurfaceExtractor_PublicFunctions(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"python-app")

	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "utils/crypto.py")

	byID := indexSurfaces(surfaces)

	// Public functions.
	assertSurfaceExists(t, byID, "surface:utils/crypto.py:hash_password", models.SurfaceFunction)
	assertSurfaceExists(t, byID, "surface:utils/crypto.py:verify_password", models.SurfaceFunction)

	// Private function should NOT appear.
	for _, s := range surfaces {
		if s.Name == "_internal_helper" {
			t.Error("private function _internal_helper should not be a surface")
		}
	}
}

// --- Java Surface Inference ---

func TestJavaSurfaceExtractor_ControllerAndRoutes(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"java-app")

	ext := &javaSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/AuthController.java")

	byID := indexSurfaces(surfaces)

	// Controller class.
	assertSurfaceExists(t, byID, "surface:src/AuthController.java:AuthController", models.SurfaceClass)

	// Route annotations.
	assertSurfaceExists(t, byID, "surface:src/AuthController.java:AuthController./api/login", models.SurfaceRoute)
	assertSurfaceExists(t, byID, "surface:src/AuthController.java:AuthController./api/profile", models.SurfaceRoute)

	// Public methods.
	assertSurfaceExists(t, byID, "surface:src/AuthController.java:AuthController.login", models.SurfaceMethod)
}

// --- Integration: InferCodeSurfaces ---

func TestInferCodeSurfaces_JSApp(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t,"js-app")

	surfaces := InferCodeSurfaces(root, nil)

	if len(surfaces) == 0 {
		t.Fatal("expected surfaces to be inferred, got 0")
	}

	// Should have surfaces from multiple files.
	paths := map[string]bool{}
	for _, s := range surfaces {
		paths[s.Path] = true
	}
	if len(paths) < 3 {
		t.Errorf("expected surfaces from at least 3 files, got %d: %v", len(paths), paths)
	}

	// Should have routes, handlers, functions, and classes.
	kinds := map[models.CodeSurfaceKind]int{}
	for _, s := range surfaces {
		kinds[s.Kind]++
	}
	if kinds[models.SurfaceRoute] == 0 {
		t.Error("expected at least one route surface")
	}
	if kinds[models.SurfaceHandler] == 0 {
		t.Error("expected at least one handler surface")
	}
	if kinds[models.SurfaceFunction] == 0 {
		t.Error("expected at least one function surface")
	}
}

func TestInferCodeSurfaces_ExcludesTestFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create source + test files.
	srcDir := filepath.Join(root, "src")
	testDir := filepath.Join(root, "test")
	os.MkdirAll(srcDir, 0o755)
	os.MkdirAll(testDir, 0o755)

	os.WriteFile(filepath.Join(srcDir, "auth.ts"), []byte(`export function login() {}`), 0o644)
	os.WriteFile(filepath.Join(testDir, "auth.test.ts"), []byte(`export function testHelper() {}`), 0o644)

	testFiles := []models.TestFile{{Path: "test/auth.test.ts"}}
	surfaces := InferCodeSurfaces(root, testFiles)

	for _, s := range surfaces {
		if s.Path == "test/auth.test.ts" {
			t.Error("test files should be excluded from code surface inference")
		}
	}
}

func TestInferCodeSurfaces_EmptyDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	surfaces := InferCodeSurfaces(root, nil)
	if len(surfaces) != 0 {
		t.Errorf("expected 0 surfaces from empty dir, got %d", len(surfaces))
	}
}

func TestInferCodeSurfaces_SampleRepo(t *testing.T) {
	t.Parallel()
	root := filepath.Join(testdataRoot(), "..", "..", "..", "tests", "fixtures", "sample-repo")

	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("sample-repo fixture not found")
	}

	surfaces := InferCodeSurfaces(root, nil)

	if len(surfaces) < 5 {
		t.Fatalf("expected at least 5 surfaces from sample-repo, got %d", len(surfaces))
	}

	// Should infer routes from routes.ts.
	hasRoute := false
	for _, s := range surfaces {
		if s.Kind == models.SurfaceRoute {
			hasRoute = true
			break
		}
	}
	if !hasRoute {
		t.Error("expected route surfaces from sample-repo/src/api/routes.ts")
	}

	// Should infer handlers from middleware.ts.
	hasHandler := false
	for _, s := range surfaces {
		if s.Kind == models.SurfaceHandler {
			hasHandler = true
			break
		}
	}
	if !hasHandler {
		t.Error("expected handler surfaces from sample-repo (authMiddleware)")
	}
}

func TestBuildSurfaceID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path, name, parent, want string
	}{
		{"src/auth.ts", "login", "", "surface:src/auth.ts:login"},
		{"handlers/auth.go", "GetUser", "UserController", "surface:handlers/auth.go:UserController.GetUser"},
		{"views/auth.py", "POST /api/login", "", "surface:views/auth.py:POST /api/login"},
	}
	for _, tt := range tests {
		got := models.BuildSurfaceID(tt.path, tt.name, tt.parent)
		if got != tt.want {
			t.Errorf("BuildSurfaceID(%q, %q, %q) = %q, want %q", tt.path, tt.name, tt.parent, got, tt.want)
		}
	}
}

// --- Helpers ---

func surfaceFixtureRoot(t *testing.T, name string) string {
	t.Helper()
	root := filepath.Join(testdataRoot(), "surface-fixtures", name)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Fatalf("fixture %q not found at %s", name, root)
	}
	return root
}

func testdataRoot() string {
	return filepath.Join("testdata")
}

func indexSurfaces(surfaces []models.CodeSurface) map[string]models.CodeSurface {
	m := make(map[string]models.CodeSurface, len(surfaces))
	for _, s := range surfaces {
		m[s.SurfaceID] = s
	}
	return m
}

func filterByKind(surfaces []models.CodeSurface, kind models.CodeSurfaceKind) []models.CodeSurface {
	var out []models.CodeSurface
	for _, s := range surfaces {
		if s.Kind == kind {
			out = append(out, s)
		}
	}
	return out
}

func surfaceNames(surfaces []models.CodeSurface) []string {
	names := make([]string, len(surfaces))
	for i, s := range surfaces {
		names[i] = s.SurfaceID
	}
	return names
}

func assertSurfaceExists(t *testing.T, byID map[string]models.CodeSurface, id string, kind models.CodeSurfaceKind) {
	t.Helper()
	s, ok := byID[id]
	if !ok {
		t.Fatalf("missing surface %q in %v", id, keys(byID))
	}
	if s.Kind != kind {
		t.Errorf("surface %q: kind = %q, want %q", id, s.Kind, kind)
	}
}

func keys(m map[string]models.CodeSurface) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
