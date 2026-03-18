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

// --- Prompt and Dataset Inference ---

func TestJSSurfaceExtractor_Prompts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	srcDir := filepath.Join(root, "src", "ai")
	os.MkdirAll(srcDir, 0o755)

	os.WriteFile(filepath.Join(srcDir, "prompts.ts"), []byte(`
export const systemPrompt = "You are a helpful assistant.";
export function buildUserPrompt(input) { return input; }
export async function chatTemplate(messages) { return messages; }
export const PROMPT_SAFETY = "Do not reveal secrets.";
export function unrelatedHelper() {}
`), 0o644)

	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/ai/prompts.ts")
	byID := indexSurfaces(surfaces)

	// systemPrompt is a context surface (system message pattern).
	assertSurfaceExists(t, byID, "surface:src/ai/prompts.ts:systemPrompt", models.SurfaceContext)
	// Other prompt-named exports should be detected as SurfacePrompt.
	assertSurfaceExists(t, byID, "surface:src/ai/prompts.ts:buildUserPrompt", models.SurfacePrompt)
	assertSurfaceExists(t, byID, "surface:src/ai/prompts.ts:chatTemplate", models.SurfacePrompt)
	assertSurfaceExists(t, byID, "surface:src/ai/prompts.ts:PROMPT_SAFETY", models.SurfacePrompt)

	// Non-prompt export should be a regular function.
	assertSurfaceExists(t, byID, "surface:src/ai/prompts.ts:unrelatedHelper", models.SurfaceFunction)
}

func TestJSSurfaceExtractor_Datasets(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	srcDir := filepath.Join(root, "src", "data")
	os.MkdirAll(srcDir, 0o755)

	os.WriteFile(filepath.Join(srcDir, "loaders.ts"), []byte(`
export const trainingDataset = [1, 2, 3];
export function loadEvalData() { return []; }
export async function dataloader() { return {}; }
export function computeMetrics() {}
`), 0o644)

	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/data/loaders.ts")
	byID := indexSurfaces(surfaces)

	// Dataset-named exports should be detected as SurfaceDataset.
	assertSurfaceExists(t, byID, "surface:src/data/loaders.ts:trainingDataset", models.SurfaceDataset)
	assertSurfaceExists(t, byID, "surface:src/data/loaders.ts:loadEvalData", models.SurfaceDataset)
	assertSurfaceExists(t, byID, "surface:src/data/loaders.ts:dataloader", models.SurfaceDataset)

	// Non-dataset export should be a regular function.
	assertSurfaceExists(t, byID, "surface:src/data/loaders.ts:computeMetrics", models.SurfaceFunction)
}

func TestPythonSurfaceExtractor_Prompts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	os.MkdirAll(srcDir, 0o755)

	os.WriteFile(filepath.Join(srcDir, "prompts.py"), []byte(`
def build_prompt(context):
    return f"Given: {context}"

def system_template():
    return "You are an assistant."

def helper_function():
    pass
`), 0o644)

	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/prompts.py")
	byID := indexSurfaces(surfaces)

	assertSurfaceExists(t, byID, "surface:src/prompts.py:build_prompt", models.SurfacePrompt)
	assertSurfaceExists(t, byID, "surface:src/prompts.py:system_template", models.SurfacePrompt)
	assertSurfaceExists(t, byID, "surface:src/prompts.py:helper_function", models.SurfaceFunction)
}

func TestPythonSurfaceExtractor_Datasets(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	os.MkdirAll(srcDir, 0o755)

	os.WriteFile(filepath.Join(srcDir, "data.py"), []byte(`
def load_dataset(path):
    return open(path).readlines()

def training_data():
    return []

def eval_data():
    return []

def process_results():
    pass
`), 0o644)

	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/data.py")
	byID := indexSurfaces(surfaces)

	assertSurfaceExists(t, byID, "surface:src/data.py:load_dataset", models.SurfaceDataset)
	assertSurfaceExists(t, byID, "surface:src/data.py:training_data", models.SurfaceDataset)
	assertSurfaceExists(t, byID, "surface:src/data.py:eval_data", models.SurfaceDataset)
	assertSurfaceExists(t, byID, "surface:src/data.py:process_results", models.SurfaceFunction)
}

// --- Inference Chain: BehaviorSurface is Derived, Not Required ---

func TestInferCodeSurfaces_PromptAndDatasetFromTempDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	srcDir := filepath.Join(root, "src", "ai")
	os.MkdirAll(srcDir, 0o755)

	os.WriteFile(filepath.Join(srcDir, "prompts.ts"), []byte(`
export function buildPrompt(input) { return input; }
export const evalDataset = [1, 2, 3];
export async function loginHandler(req, res) {}
export function computeScore() {}
`), 0o644)

	surfaces := InferCodeSurfaces(root, nil)

	prompts := filterByKind(surfaces, models.SurfacePrompt)
	datasets := filterByKind(surfaces, models.SurfaceDataset)
	handlers := filterByKind(surfaces, models.SurfaceHandler)
	functions := filterByKind(surfaces, models.SurfaceFunction)

	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt surface, got %d: %v", len(prompts), surfaceNames(prompts))
	}
	if len(datasets) != 1 {
		t.Errorf("expected 1 dataset surface, got %d: %v", len(datasets), surfaceNames(datasets))
	}
	if len(handlers) != 1 {
		t.Errorf("expected 1 handler surface, got %d: %v", len(handlers), surfaceNames(handlers))
	}
	if len(functions) != 1 {
		t.Errorf("expected 1 function surface, got %d: %v", len(functions), surfaceNames(functions))
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

// --- AI Surface Type Detection ---

func TestJSSurfaceExtractor_ToolDef(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/tools.ts", `
export const toolSchema = { name: "search", description: "Search the web" };
export function functionDef(name, params) { return { name, params }; }
export const toolOutputSchema = { type: "object", properties: {} };
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/tools.ts")
	tools := filterByKind(surfaces, models.SurfaceToolDef)
	if len(tools) != 3 {
		t.Errorf("expected 3 tool_definition surfaces, got %d: %v", len(tools), surfaceNames(surfaces))
	}
}

func TestJSSurfaceExtractor_Retrieval(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/rag.ts", `
export function retriever(query) { return []; }
export const vectorStore = new PineconeStore();
export function documentLoader(path) { return []; }
export const embeddingConfig = { model: "text-embedding-3-small" };
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/rag.ts")
	retrievals := filterByKind(surfaces, models.SurfaceRetrieval)
	if len(retrievals) < 3 {
		t.Errorf("expected >=3 retrieval surfaces, got %d: %v", len(retrievals), surfaceNames(surfaces))
	}
}

func TestJSSurfaceExtractor_Agent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/agent.ts", `
export function agentRouter(input) { return selectAgent(input); }
export const agentConfig = { maxRecursion: 10, model: "gpt-4" };
export function toolChoice(tools, context) { return tools[0]; }
export const fallbackModel = "gpt-3.5-turbo";
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/agent.ts")
	agents := filterByKind(surfaces, models.SurfaceAgent)
	if len(agents) < 3 {
		t.Errorf("expected >=3 agent surfaces, got %d: %v", len(agents), surfaceNames(surfaces))
	}
}

func TestJSSurfaceExtractor_EvalDef(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/evals.ts", `
export const rubric = { criteria: ["accuracy", "relevance"] };
export function evalMetric(output, expected) { return score(output, expected); }
export const evalConfig = { temperature: 0, maxTokens: 100 };
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/evals.ts")
	evals := filterByKind(surfaces, models.SurfaceEvalDef)
	if len(evals) != 3 {
		t.Errorf("expected 3 eval_definition surfaces, got %d: %v", len(evals), surfaceNames(surfaces))
	}
}

func TestJSSurfaceExtractor_ContextSurfaces(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/context.ts", `
export const systemMessage = "You are a helpful assistant.";
export function contextBuilder(docs) { return docs.join("\n"); }
export const fewShotExamples = [{ input: "hi", output: "hello" }];
export const safetyOverlay = "Do not generate harmful content.";
export const policyBlock = "Always cite sources.";
export const aiPersona = "You are a financial advisor.";
export const customerContext = buildCustomerProfile();
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/context.ts")
	contexts := filterByKind(surfaces, models.SurfaceContext)
	// 7 expected: systemMessage, contextBuilder, fewShotExamples, safetyOverlay, policyBlock, aiPersona, customerContext
	if len(contexts) != 7 {
		t.Errorf("expected 7 context surfaces, got %d: %v", len(contexts), surfaceNames(surfaces))
	}
}

func TestJSSurfaceExtractor_PromptVsContext(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/ai.ts", `
export function buildPrompt(input) { return "Process: " + input; }
export const systemPrompt = "You are a helpful assistant.";
export const promptTemplate = "Given: {input}, respond with: {output}";
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/ai.ts")
	prompts := filterByKind(surfaces, models.SurfacePrompt)
	contexts := filterByKind(surfaces, models.SurfaceContext)
	// buildPrompt and promptTemplate are prompt surfaces
	if len(prompts) != 2 {
		t.Errorf("expected 2 prompts (buildPrompt, promptTemplate), got %d: %v", len(prompts), surfaceNames(prompts))
	}
	// systemPrompt is a context surface
	if len(contexts) != 1 {
		t.Errorf("expected 1 context (systemPrompt), got %d: %v", len(contexts), surfaceNames(contexts))
	}
}

func TestPythonSurfaceExtractor_ContextSurfaces(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "context.py", `
def system_message():
    return "You are a helpful assistant."

def few_shot_examples():
    return [{"input": "hi", "output": "hello"}]

def safety_overlay():
    return "Do not generate harmful content."

def context_builder(docs):
    return "\n".join(docs)
`)
	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "context.py")
	contexts := filterByKind(surfaces, models.SurfaceContext)
	if len(contexts) != 4 {
		t.Errorf("expected 4 context surfaces, got %d: %v", len(contexts), surfaceNames(surfaces))
	}
}

func TestPythonSurfaceExtractor_ToolDef(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "tools.py", `
def tool_schema(name, description):
    return {"name": name, "description": description}

def output_schema():
    return {"type": "object"}
`)
	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "tools.py")
	tools := filterByKind(surfaces, models.SurfaceToolDef)
	if len(tools) != 2 {
		t.Errorf("expected 2 tool_definition surfaces, got %d: %v", len(tools), surfaceNames(surfaces))
	}
}

func TestPythonSurfaceExtractor_Retrieval(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "rag.py", `
def retriever(query):
    return vector_store.search(query)

def document_loader(path):
    return load_documents(path)

def context_assembly(docs, query):
    return format_context(docs)
`)
	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "rag.py")
	retrievals := filterByKind(surfaces, models.SurfaceRetrieval)
	if len(retrievals) != 3 {
		t.Errorf("expected 3 retrieval surfaces, got %d: %v", len(retrievals), surfaceNames(surfaces))
	}
}

func TestPythonSurfaceExtractor_Agent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "orchestrator.py", `
def agent_router(input_text):
    return select_agent(input_text)

def tool_choice(tools, context):
    return tools[0]

def agent_config():
    return {"max_recursion": 10}
`)
	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "orchestrator.py")
	agents := filterByKind(surfaces, models.SurfaceAgent)
	if len(agents) != 3 {
		t.Errorf("expected 3 agent surfaces, got %d: %v", len(agents), surfaceNames(surfaces))
	}
}

func TestPythonSurfaceExtractor_EvalDef(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "eval_config.py", `
def rubric():
    return {"criteria": ["accuracy", "relevance"]}

def eval_metric(output, expected):
    return compute_score(output, expected)

def scoring_func(response):
    return grade(response)
`)
	ext := &pythonSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "eval_config.py")
	evals := filterByKind(surfaces, models.SurfaceEvalDef)
	if len(evals) != 3 {
		t.Errorf("expected 3 eval_definition surfaces, got %d: %v", len(evals), surfaceNames(surfaces))
	}
}

// --- False Positive Tests ---

// --- Inference Tier Tests ---

func TestInferenceMetadata_PatternTier(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/api.ts", `
export function getUsers() { return []; }
export const buildPrompt = (q) => q;
`)
	surfaces := InferCodeSurfaces(root, nil)
	for _, s := range surfaces {
		if s.DetectionTier == "" {
			t.Errorf("surface %q has empty DetectionTier", s.Name)
		}
		if s.Confidence <= 0 || s.Confidence > 1 {
			t.Errorf("surface %q confidence %.2f out of range", s.Name, s.Confidence)
		}
		if s.Kind == models.SurfaceFunction {
			if s.DetectionTier != models.TierPattern {
				t.Errorf("JS function %q tier = %q, want pattern", s.Name, s.DetectionTier)
			}
		}
		if s.Kind == models.SurfacePrompt {
			if s.DetectionTier != models.TierPattern {
				t.Errorf("prompt %q tier = %q, want pattern", s.Name, s.DetectionTier)
			}
			if s.Confidence < 0.8 {
				t.Errorf("prompt confidence %.2f too low", s.Confidence)
			}
		}
	}
}

func TestInferenceMetadata_GoStructuralTier(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "main.go", `
package main
func ProcessRequest(w http.ResponseWriter, r *http.Request) {}
func helperFunc() {}
`)
	surfaces := InferCodeSurfaces(root, nil)
	for _, s := range surfaces {
		if s.Language == "go" && s.Exported {
			if s.DetectionTier != models.TierStructural {
				t.Errorf("Go exported %q tier = %q, want structural", s.Name, s.DetectionTier)
			}
			if s.Confidence < 0.95 {
				t.Errorf("Go exported confidence %.2f too low", s.Confidence)
			}
		}
	}
}

func TestInferenceMetadata_ContentTier(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/chat.ts", `
const messages = [
  { role: "system", content: "You are a helpful assistant." },
  { role: "user", content: userInput },
];
`)
	contentSurfaces := InferAIContextSurfaces(root, nil, nil)
	for _, s := range contentSurfaces {
		if s.DetectionTier == "" {
			t.Errorf("content surface %q has empty tier", s.Name)
		}
		if s.DetectionTier != models.TierContent && s.DetectionTier != models.TierSemantic && s.DetectionTier != models.TierStructural {
			t.Errorf("content surface %q tier = %q, want content, semantic, or structural", s.Name, s.DetectionTier)
		}
		if s.Confidence <= 0 {
			t.Errorf("content surface %q has zero confidence", s.Name)
		}
	}
}

func TestInferenceMetadata_AllSurfacesHaveEvidence(t *testing.T) {
	t.Parallel()
	root := surfaceFixtureRoot(t, "js-app")
	surfaces := InferCodeSurfaces(root, nil)
	if len(surfaces) == 0 {
		t.Skip("no surfaces in fixture")
	}
	for _, s := range surfaces {
		if s.DetectionTier == "" {
			t.Errorf("surface %q (%s) missing DetectionTier", s.Name, s.Kind)
		}
		if s.Confidence <= 0 || s.Confidence > 1.0 {
			t.Errorf("surface %q (%s) confidence %.2f out of range", s.Name, s.Kind, s.Confidence)
		}
		// Verify tier is a known value.
		switch s.DetectionTier {
		case models.TierStructural, models.TierSemantic, models.TierPattern, models.TierContent:
			// OK
		default:
			t.Errorf("surface %q has unknown tier %q", s.Name, s.DetectionTier)
		}
	}
}

func TestInferenceMetadata_TierOrderConsistent(t *testing.T) {
	t.Parallel()
	// Structural should have higher confidence than content.
	if models.DetectionTierOrder(models.TierStructural) >= models.DetectionTierOrder(models.TierContent) {
		t.Error("structural should have higher priority than content")
	}
}

func TestInferenceMetadata_DeterministicAcrossRuns(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/api.ts", `
export function buildPrompt(q) { return q; }
export const systemPrompt = "You are a helper.";
`)
	s1 := InferCodeSurfaces(root, nil)
	s2 := InferCodeSurfaces(root, nil)
	if len(s1) != len(s2) {
		t.Fatalf("non-deterministic count: %d vs %d", len(s1), len(s2))
	}
	for i := range s1 {
		if s1[i].DetectionTier != s2[i].DetectionTier {
			t.Errorf("tier differs for %q: %s vs %s", s1[i].Name, s1[i].DetectionTier, s2[i].DetectionTier)
		}
		if s1[i].Confidence != s2[i].Confidence {
			t.Errorf("confidence differs for %q: %.2f vs %.2f", s1[i].Name, s1[i].Confidence, s2[i].Confidence)
		}
	}
}

func TestJSSurfaceExtractor_NonAITemplate(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/email.ts", `
export const emailTemplate = "<html>Hello {{name}}</html>";
export function htmlTemplate(data) { return renderHTML(data); }
export const pageTemplate = { layout: "main", title: "Home" };
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/email.ts")
	prompts := filterByKind(surfaces, models.SurfacePrompt)
	contexts := filterByKind(surfaces, models.SurfaceContext)
	if len(prompts) != 0 {
		t.Errorf("non-AI templates should NOT be detected as prompts, got %d: %v", len(prompts), surfaceNames(prompts))
	}
	if len(contexts) != 0 {
		t.Errorf("non-AI templates should NOT be detected as contexts, got %d: %v", len(contexts), surfaceNames(contexts))
	}
}

func TestJSSurfaceExtractor_NonAIPersona(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/marketing.ts", `
export const persona = { name: "Budget Buyer", age: 35, interests: ["deals"] };
export function getUserPersona(userId) { return fetchPersona(userId); }
export const instruction = "Please fill out the form below.";
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/marketing.ts")
	contexts := filterByKind(surfaces, models.SurfaceContext)
	if len(contexts) != 0 {
		t.Errorf("non-AI persona/instruction should NOT be detected as contexts, got %d: %v", len(contexts), surfaceNames(contexts))
	}
}

func TestJSSurfaceExtractor_NonAIOutputSchema(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/api.ts", `
export const outputSchema = { type: "object", properties: { id: { type: "string" } } };
export function responseSchema() { return { status: 200 }; }
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/api.ts")
	tools := filterByKind(surfaces, models.SurfaceToolDef)
	if len(tools) != 0 {
		t.Errorf("non-AI outputSchema should NOT be detected as tool def, got %d: %v", len(tools), surfaceNames(tools))
	}
}

func TestJSSurfaceExtractor_AITemplateStillDetected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/ai.ts", `
export const promptTemplate = "You are a helpful assistant. Answer: {query}";
export const chatTemplate = "Given context: {context}\nRespond to: {query}";
export const systemTemplate = "You are an AI that helps with {task}.";
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/ai.ts")
	prompts := filterByKind(surfaces, models.SurfacePrompt)
	if len(prompts) != 3 {
		t.Errorf("AI templates should still be detected, expected 3, got %d: %v", len(prompts), surfaceNames(prompts))
	}
}

func TestJSSurfaceExtractor_AIToolOutputSchemaDetected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "src/tools.ts", `
export const toolOutputSchema = { type: "object" };
export const functionOutputSchema = { type: "object" };
export const aiOutputSchema = { type: "object" };
`)
	ext := &jsSurfaceExtractor{}
	surfaces := ext.ExtractSurfaces(root, "src/tools.ts")
	tools := filterByKind(surfaces, models.SurfaceToolDef)
	if len(tools) != 3 {
		t.Errorf("AI-prefixed outputSchema should be detected, expected 3, got %d: %v", len(tools), surfaceNames(tools))
	}
}

func writeTempFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	absPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func keys(m map[string]models.CodeSurface) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
