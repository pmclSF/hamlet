package analysis

import (
	"sort"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestDeriveBehaviorSurfaces_Empty(t *testing.T) {
	t.Parallel()
	result := DeriveBehaviorSurfaces(nil)
	if len(result) != 0 {
		t.Errorf("expected 0 behaviors from nil input, got %d", len(result))
	}
}

func TestDeriveBehaviorSurfaces_RouteGrouping(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{
			SurfaceID:  "surface:src/routes/api.ts:GET /api/users",
			Name:       "GET /api/users",
			Path:       "src/routes/api.ts",
			Kind:       models.SurfaceRoute,
			Language:   "typescript",
			HTTPMethod: "GET",
			Route:      "/api/users",
			Exported:   true,
		},
		{
			SurfaceID:  "surface:src/routes/api.ts:POST /api/users",
			Name:       "POST /api/users",
			Path:       "src/routes/api.ts",
			Kind:       models.SurfaceRoute,
			Language:   "typescript",
			HTTPMethod: "POST",
			Route:      "/api/users",
			Exported:   true,
		},
		{
			SurfaceID:  "surface:src/routes/api.ts:DELETE /api/users/:id",
			Name:       "DELETE /api/users/:id",
			Path:       "src/routes/api.ts",
			Kind:       models.SurfaceRoute,
			Language:   "typescript",
			HTTPMethod: "DELETE",
			Route:      "/api/users/:id",
			Exported:   true,
		},
		{
			SurfaceID:  "surface:src/routes/api.ts:POST /api/login",
			Name:       "POST /api/login",
			Path:       "src/routes/api.ts",
			Kind:       models.SurfaceRoute,
			Language:   "typescript",
			HTTPMethod: "POST",
			Route:      "/api/login",
			Exported:   true,
		},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)

	// Should produce a route group for /api/users (3 routes share this prefix).
	byID := indexBehaviors(behaviors)

	routeBehavior, ok := byID["behavior:route:/api/users"]
	if !ok {
		t.Fatalf("expected route group /api/users, got behaviors: %v", behaviorIDs(behaviors))
	}
	if routeBehavior.Kind != models.BehaviorGroupRoutePrefix {
		t.Errorf("expected kind route_prefix, got %q", routeBehavior.Kind)
	}
	if len(routeBehavior.CodeSurfaceIDs) != 3 {
		t.Errorf("expected 3 surfaces in /api/users group, got %d", len(routeBehavior.CodeSurfaceIDs))
	}
	if routeBehavior.RoutePrefix != "/api/users" {
		t.Errorf("expected route prefix /api/users, got %q", routeBehavior.RoutePrefix)
	}
	if routeBehavior.Language != "typescript" {
		t.Errorf("expected language typescript, got %q", routeBehavior.Language)
	}
}

func TestDeriveBehaviorSurfaces_ClassGrouping(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{
			SurfaceID:  "surface:handlers/auth.go:UserController.GetUser",
			Name:       "GetUser",
			Path:       "handlers/auth.go",
			Kind:       models.SurfaceMethod,
			ParentName: "UserController",
			Language:   "go",
			Receiver:   "UserController",
			Exported:   true,
		},
		{
			SurfaceID:  "surface:handlers/auth.go:UserController.DeleteUser",
			Name:       "DeleteUser",
			Path:       "handlers/auth.go",
			Kind:       models.SurfaceMethod,
			ParentName: "UserController",
			Language:   "go",
			Receiver:   "UserController",
			Exported:   true,
		},
		{
			SurfaceID:  "surface:handlers/auth.go:UserController.UpdateUser",
			Name:       "UpdateUser",
			Path:       "handlers/auth.go",
			Kind:       models.SurfaceMethod,
			ParentName: "UserController",
			Language:   "go",
			Receiver:   "UserController",
			Exported:   true,
		},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)

	byID := indexBehaviors(behaviors)

	classBehavior, ok := byID["behavior:class:handlers/auth.go:UserController"]
	if !ok {
		t.Fatalf("expected class group UserController, got behaviors: %v", behaviorIDs(behaviors))
	}
	if classBehavior.Kind != models.BehaviorGroupClass {
		t.Errorf("expected kind class, got %q", classBehavior.Kind)
	}
	if len(classBehavior.CodeSurfaceIDs) != 3 {
		t.Errorf("expected 3 surfaces in UserController group, got %d", len(classBehavior.CodeSurfaceIDs))
	}
	if classBehavior.Label != "UserController" {
		t.Errorf("expected label UserController, got %q", classBehavior.Label)
	}
}

func TestDeriveBehaviorSurfaces_ModuleGrouping(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{
			SurfaceID: "surface:src/utils/validation.ts:validateEmail",
			Name:      "validateEmail",
			Path:      "src/utils/validation.ts",
			Kind:      models.SurfaceFunction,
			Language:  "typescript",
			Exported:  true,
		},
		{
			SurfaceID: "surface:src/utils/validation.ts:sanitizeInput",
			Name:      "sanitizeInput",
			Path:      "src/utils/validation.ts",
			Kind:      models.SurfaceFunction,
			Language:  "typescript",
			Exported:  true,
		},
		{
			SurfaceID: "surface:src/utils/validation.ts:Validator",
			Name:      "Validator",
			Path:      "src/utils/validation.ts",
			Kind:      models.SurfaceClass,
			Language:  "typescript",
			Exported:  true,
		},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)

	byID := indexBehaviors(behaviors)

	moduleBehavior, ok := byID["behavior:module:src/utils/validation.ts"]
	if !ok {
		t.Fatalf("expected module group for validation.ts, got behaviors: %v", behaviorIDs(behaviors))
	}
	if moduleBehavior.Kind != models.BehaviorGroupModule {
		t.Errorf("expected kind module, got %q", moduleBehavior.Kind)
	}
	if len(moduleBehavior.CodeSurfaceIDs) != 3 {
		t.Errorf("expected 3 surfaces in module group, got %d", len(moduleBehavior.CodeSurfaceIDs))
	}
	if moduleBehavior.Label != "validation" {
		t.Errorf("expected label 'validation', got %q", moduleBehavior.Label)
	}
}

func TestDeriveBehaviorSurfaces_ChangedFunctionMapsToSurfaces(t *testing.T) {
	t.Parallel()

	// Simulate: a developer changed loginHandler. We should be able to find
	// which behavior surfaces it belongs to, and trace back to concrete
	// code surfaces.
	surfaces := []models.CodeSurface{
		{
			SurfaceID:  "surface:src/routes/auth.ts:POST /api/login",
			Name:       "POST /api/login",
			Path:       "src/routes/auth.ts",
			Kind:       models.SurfaceRoute,
			Language:   "typescript",
			HTTPMethod: "POST",
			Route:      "/api/login",
			Exported:   true,
		},
		{
			SurfaceID:  "surface:src/routes/auth.ts:POST /api/register",
			Name:       "POST /api/register",
			Path:       "src/routes/auth.ts",
			Kind:       models.SurfaceRoute,
			Language:   "typescript",
			HTTPMethod: "POST",
			Route:      "/api/register",
			Exported:   true,
		},
		{
			SurfaceID: "surface:src/routes/auth.ts:loginHandler",
			Name:      "loginHandler",
			Path:      "src/routes/auth.ts",
			Kind:      models.SurfaceHandler,
			Language:  "typescript",
			Exported:  true,
		},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)

	// The changed function "loginHandler" has surface ID
	// "surface:src/routes/auth.ts:loginHandler". Find which behaviors
	// contain it.
	changedSurfaceID := "surface:src/routes/auth.ts:loginHandler"
	var containingBehaviors []models.BehaviorSurface
	for _, b := range behaviors {
		for _, csID := range b.CodeSurfaceIDs {
			if csID == changedSurfaceID {
				containingBehaviors = append(containingBehaviors, b)
				break
			}
		}
	}

	if len(containingBehaviors) == 0 {
		t.Fatal("changed function loginHandler should map to at least one behavior surface")
	}

	// Every containing behavior should reference concrete code surface IDs.
	for _, b := range containingBehaviors {
		if len(b.CodeSurfaceIDs) == 0 {
			t.Errorf("behavior %q has no code surface references", b.BehaviorID)
		}
		// Verify the explanation references concrete origins.
		if b.Description == "" {
			t.Errorf("behavior %q has no description explaining its derivation", b.BehaviorID)
		}
	}
}

func TestDeriveBehaviorSurfaces_ExplainOutputReferencesCodeSurfaces(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{
			SurfaceID:  "surface:views/auth.py:POST /api/login",
			Name:       "POST /api/login",
			Path:       "views/auth.py",
			Kind:       models.SurfaceRoute,
			Language:   "python",
			HTTPMethod: "POST",
			Route:      "/api/login",
		},
		{
			SurfaceID:  "surface:views/auth.py:GET /api/profile",
			Name:       "GET /api/profile",
			Path:       "views/auth.py",
			Kind:       models.SurfaceRoute,
			Language:   "python",
			HTTPMethod: "GET",
			Route:      "/api/profile",
		},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)

	for _, b := range behaviors {
		// Every behavior must trace back to concrete code surfaces.
		if len(b.CodeSurfaceIDs) == 0 {
			t.Errorf("behavior %q has no code surface references — explanations must be grounded", b.BehaviorID)
		}
		// All referenced code surface IDs must exist in the input.
		surfaceSet := map[string]bool{}
		for _, s := range surfaces {
			surfaceSet[s.SurfaceID] = true
		}
		for _, csID := range b.CodeSurfaceIDs {
			if !surfaceSet[csID] {
				t.Errorf("behavior %q references unknown code surface %q", b.BehaviorID, csID)
			}
		}
	}
}

func TestDeriveBehaviorSurfaces_SingleSurfaceNoBehavior(t *testing.T) {
	t.Parallel()
	// A single surface shouldn't produce any behavior groups.
	surfaces := []models.CodeSurface{
		{
			SurfaceID: "surface:src/index.ts:main",
			Name:      "main",
			Path:      "src/index.ts",
			Kind:      models.SurfaceFunction,
			Language:  "typescript",
			Exported:  true,
		},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)

	if len(behaviors) != 0 {
		t.Errorf("expected 0 behaviors from single surface, got %d: %v",
			len(behaviors), behaviorIDs(behaviors))
	}
}

func TestDeriveBehaviorSurfaces_DeterministicOutput(t *testing.T) {
	t.Parallel()
	surfaces := []models.CodeSurface{
		{SurfaceID: "surface:a.ts:foo", Name: "foo", Path: "a.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
		{SurfaceID: "surface:a.ts:bar", Name: "bar", Path: "a.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
		{SurfaceID: "surface:a.ts:baz", Name: "baz", Path: "a.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
	}

	// Run derivation twice and ensure identical output.
	b1 := DeriveBehaviorSurfaces(surfaces)
	b2 := DeriveBehaviorSurfaces(surfaces)

	if len(b1) != len(b2) {
		t.Fatalf("non-deterministic: got %d then %d behaviors", len(b1), len(b2))
	}
	for i := range b1 {
		if b1[i].BehaviorID != b2[i].BehaviorID {
			t.Errorf("non-deterministic at index %d: %q vs %q", i, b1[i].BehaviorID, b2[i].BehaviorID)
		}
		// Also check code surface ID ordering.
		for j := range b1[i].CodeSurfaceIDs {
			if b1[i].CodeSurfaceIDs[j] != b2[i].CodeSurfaceIDs[j] {
				t.Errorf("non-deterministic surface IDs at behavior %d, index %d", i, j)
			}
		}
	}
}

func TestDeriveBehaviorSurfaces_MixedStrategies(t *testing.T) {
	t.Parallel()

	// A realistic mix: routes, class methods, and standalone exports.
	surfaces := []models.CodeSurface{
		// Routes under /api/auth
		{SurfaceID: "surface:src/auth.ts:POST /api/auth/login", Name: "POST /api/auth/login", Path: "src/auth.ts", Kind: models.SurfaceRoute, Language: "typescript", Route: "/api/auth/login"},
		{SurfaceID: "surface:src/auth.ts:POST /api/auth/register", Name: "POST /api/auth/register", Path: "src/auth.ts", Kind: models.SurfaceRoute, Language: "typescript", Route: "/api/auth/register"},
		// Class methods
		{SurfaceID: "surface:src/auth.ts:AuthService.validate", Name: "validate", Path: "src/auth.ts", Kind: models.SurfaceMethod, ParentName: "AuthService", Language: "typescript"},
		{SurfaceID: "surface:src/auth.ts:AuthService.refresh", Name: "refresh", Path: "src/auth.ts", Kind: models.SurfaceMethod, ParentName: "AuthService", Language: "typescript"},
		// Standalone export
		{SurfaceID: "surface:src/auth.ts:hashPassword", Name: "hashPassword", Path: "src/auth.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)

	// Should have: route group (/api/auth), class group (AuthService),
	// module group (src/auth.ts).
	byID := indexBehaviors(behaviors)

	if _, ok := byID["behavior:route:/api/auth"]; !ok {
		t.Errorf("expected route group /api/auth, got: %v", behaviorIDs(behaviors))
	}
	if _, ok := byID["behavior:class:src/auth.ts:AuthService"]; !ok {
		t.Errorf("expected class group AuthService, got: %v", behaviorIDs(behaviors))
	}
	if _, ok := byID["behavior:module:src/auth.ts"]; !ok {
		t.Errorf("expected module group src/auth.ts, got: %v", behaviorIDs(behaviors))
	}
}

func TestDeriveBehaviorSurfaces_IntegrationWithFixtures(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t, "js-app")

	// First infer code surfaces, then derive behaviors from them.
	codeSurfaces := InferCodeSurfaces(root, nil)
	if len(codeSurfaces) == 0 {
		t.Fatal("expected code surfaces from js-app fixture")
	}

	behaviors := DeriveBehaviorSurfaces(codeSurfaces)

	// js-app has routes in api.ts, handlers in auth.ts, exports in validation.ts.
	// Should produce at least a module group.
	if len(behaviors) == 0 {
		t.Fatal("expected at least one behavior surface from js-app fixture")
	}

	// Every behavior must reference valid code surface IDs.
	surfaceSet := map[string]bool{}
	for _, cs := range codeSurfaces {
		surfaceSet[cs.SurfaceID] = true
	}
	for _, b := range behaviors {
		for _, csID := range b.CodeSurfaceIDs {
			if !surfaceSet[csID] {
				t.Errorf("behavior %q references unknown surface %q", b.BehaviorID, csID)
			}
		}
	}
}

func TestDeriveBehaviorSurfaces_DomainGrouping(t *testing.T) {
	t.Parallel()

	// Surfaces spread across multiple files under src/auth/.
	surfaces := []models.CodeSurface{
		{SurfaceID: "surface:src/auth/login.ts:login", Name: "login", Path: "src/auth/login.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
		{SurfaceID: "surface:src/auth/login.ts:loginHandler", Name: "loginHandler", Path: "src/auth/login.ts", Kind: models.SurfaceHandler, Language: "typescript", Exported: true},
		{SurfaceID: "surface:src/auth/register.ts:register", Name: "register", Path: "src/auth/register.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
		{SurfaceID: "surface:src/auth/token.ts:validateToken", Name: "validateToken", Path: "src/auth/token.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)
	byID := indexBehaviors(behaviors)

	domainBehavior, ok := byID["behavior:domain:src/auth"]
	if !ok {
		t.Fatalf("expected domain group src/auth, got behaviors: %v", behaviorIDs(behaviors))
	}
	if domainBehavior.Kind != models.BehaviorGroupDomain {
		t.Errorf("expected kind domain, got %q", domainBehavior.Kind)
	}
	if len(domainBehavior.CodeSurfaceIDs) != 4 {
		t.Errorf("expected 4 surfaces in domain group, got %d", len(domainBehavior.CodeSurfaceIDs))
	}
	if domainBehavior.Label != "auth" {
		t.Errorf("expected label 'auth', got %q", domainBehavior.Label)
	}
	if domainBehavior.Language != "typescript" {
		t.Errorf("expected language typescript, got %q", domainBehavior.Language)
	}
}

func TestDeriveBehaviorSurfaces_DomainRequiresMultipleFiles(t *testing.T) {
	t.Parallel()

	// 3 surfaces but all in one file — should NOT produce a domain group.
	surfaces := []models.CodeSurface{
		{SurfaceID: "surface:src/auth/login.ts:a", Name: "a", Path: "src/auth/login.ts", Kind: models.SurfaceFunction, Language: "typescript"},
		{SurfaceID: "surface:src/auth/login.ts:b", Name: "b", Path: "src/auth/login.ts", Kind: models.SurfaceFunction, Language: "typescript"},
		{SurfaceID: "surface:src/auth/login.ts:c", Name: "c", Path: "src/auth/login.ts", Kind: models.SurfaceFunction, Language: "typescript"},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)
	byID := indexBehaviors(behaviors)

	if _, ok := byID["behavior:domain:src/auth"]; ok {
		t.Error("domain group should require surfaces from at least 2 files")
	}
}

func TestDeriveBehaviorSurfaces_NamingGrouping(t *testing.T) {
	t.Parallel()

	// Surfaces with shared naming prefix.
	surfaces := []models.CodeSurface{
		{SurfaceID: "surface:src/auth.go:AuthLogin", Name: "AuthLogin", Path: "src/auth.go", Kind: models.SurfaceFunction, Language: "go", Exported: true},
		{SurfaceID: "surface:src/auth.go:AuthRegister", Name: "AuthRegister", Path: "src/auth.go", Kind: models.SurfaceFunction, Language: "go", Exported: true},
		{SurfaceID: "surface:src/auth.go:AuthLogout", Name: "AuthLogout", Path: "src/auth.go", Kind: models.SurfaceFunction, Language: "go", Exported: true},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)
	byID := indexBehaviors(behaviors)

	namingBehavior, ok := byID["behavior:naming:Auth"]
	if !ok {
		t.Fatalf("expected naming group Auth, got behaviors: %v", behaviorIDs(behaviors))
	}
	if namingBehavior.Kind != models.BehaviorGroupNaming {
		t.Errorf("expected kind naming, got %q", namingBehavior.Kind)
	}
	if len(namingBehavior.CodeSurfaceIDs) != 3 {
		t.Errorf("expected 3 surfaces in naming group, got %d", len(namingBehavior.CodeSurfaceIDs))
	}
	if namingBehavior.Label != "Auth*" {
		t.Errorf("expected label 'Auth*', got %q", namingBehavior.Label)
	}
}

func TestDeriveBehaviorSurfaces_NamingSnakeCase(t *testing.T) {
	t.Parallel()

	// Python-style snake_case naming with shared prefix.
	surfaces := []models.CodeSurface{
		{SurfaceID: "surface:utils/crypto.py:hash_password", Name: "hash_password", Path: "utils/crypto.py", Kind: models.SurfaceFunction, Language: "python"},
		{SurfaceID: "surface:utils/crypto.py:hash_token", Name: "hash_token", Path: "utils/crypto.py", Kind: models.SurfaceFunction, Language: "python"},
		{SurfaceID: "surface:utils/crypto.py:hash_session", Name: "hash_session", Path: "utils/crypto.py", Kind: models.SurfaceFunction, Language: "python"},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)
	byID := indexBehaviors(behaviors)

	namingBehavior, ok := byID["behavior:naming:hash"]
	if !ok {
		t.Fatalf("expected naming group hash, got behaviors: %v", behaviorIDs(behaviors))
	}
	if len(namingBehavior.CodeSurfaceIDs) != 3 {
		t.Errorf("expected 3 surfaces in hash naming group, got %d", len(namingBehavior.CodeSurfaceIDs))
	}
}

func TestDeriveBehaviorSurfaces_NamingSkipsRoutes(t *testing.T) {
	t.Parallel()

	// Route names like "GET /api/users" should NOT produce naming groups.
	surfaces := []models.CodeSurface{
		{SurfaceID: "surface:api.ts:GET /api/users", Name: "GET /api/users", Path: "api.ts", Kind: models.SurfaceRoute, Route: "/api/users"},
		{SurfaceID: "surface:api.ts:GET /api/posts", Name: "GET /api/posts", Path: "api.ts", Kind: models.SurfaceRoute, Route: "/api/posts"},
		{SurfaceID: "surface:api.ts:GET /api/comments", Name: "GET /api/comments", Path: "api.ts", Kind: models.SurfaceRoute, Route: "/api/comments"},
	}

	behaviors := DeriveBehaviorSurfaces(surfaces)

	for _, b := range behaviors {
		if b.Kind == models.BehaviorGroupNaming {
			t.Errorf("route names should not produce naming groups, got %q", b.BehaviorID)
		}
	}
}

func TestExtractNamePrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{"AuthLogin", "Auth"},
		{"authLogin", "auth"},
		{"getUserById", "get"},
		{"validateEmail", "validate"},
		{"Go", ""},
		{"ab", ""},
		{"GET /api/foo", ""},
		{"hash_password", "hash"},
		{"_internal", ""},
		{"NewUserService", "New"},
		{"simple", ""},
	}
	for _, tt := range tests {
		got := extractNamePrefix(tt.name)
		if got != tt.want {
			t.Errorf("extractNamePrefix(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// --- Helpers ---

func indexBehaviors(behaviors []models.BehaviorSurface) map[string]models.BehaviorSurface {
	m := make(map[string]models.BehaviorSurface, len(behaviors))
	for _, b := range behaviors {
		m[b.BehaviorID] = b
	}
	return m
}

func behaviorIDs(behaviors []models.BehaviorSurface) []string {
	ids := make([]string, len(behaviors))
	for i, b := range behaviors {
		ids[i] = b.BehaviorID
	}
	sort.Strings(ids)
	return ids
}
