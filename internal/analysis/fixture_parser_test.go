package analysis

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestDetectJSFixtures_LifecycleHooks(t *testing.T) {
	t.Parallel()
	src := `
import { render } from '@testing-library/react';

describe('UserComponent', () => {
  let wrapper;

  beforeEach(() => {
    wrapper = render(<User />);
  });

  afterEach(() => {
    wrapper.unmount();
  });

  beforeAll(() => {
    setupDatabase();
  });

  afterAll(() => {
    teardownDatabase();
  });

  it('should render', () => {
    expect(wrapper).toBeDefined();
  });
});
`
	fixtures := detectJSFixtures(src, "test/UserComponent.test.tsx", "jest")

	hookNames := map[string]bool{}
	for _, f := range fixtures {
		hookNames[f.Name] = true
	}

	for _, name := range []string{"beforeEach", "afterEach", "beforeAll", "afterAll"} {
		if !hookNames[name] {
			t.Errorf("expected hook %q to be detected", name)
		}
	}

	// Verify scope and kind classification.
	for _, f := range fixtures {
		switch f.Name {
		case "beforeEach":
			if f.Kind != models.FixtureSetupHook {
				t.Errorf("beforeEach: want kind setup_hook, got %s", f.Kind)
			}
			if f.Scope != "test" {
				t.Errorf("beforeEach: want scope test, got %s", f.Scope)
			}
		case "afterAll":
			if f.Kind != models.FixtureTeardownHook {
				t.Errorf("afterAll: want kind teardown_hook, got %s", f.Kind)
			}
			if f.Scope != "suite" {
				t.Errorf("afterAll: want scope suite, got %s", f.Scope)
			}
		}
	}
}

func TestDetectJSFixtures_Helpers(t *testing.T) {
	t.Parallel()
	src := `
export function createUser(overrides = {}) {
  return { id: 1, name: 'test', ...overrides };
}

const buildOrder = (items) => ({ items, total: items.length });

function setupTestEnv() {
  process.env.NODE_ENV = 'test';
}

const mockAuthService = jest.fn();
`
	fixtures := detectJSFixtures(src, "test/helpers.ts", "jest")

	names := map[string]models.FixtureKind{}
	for _, f := range fixtures {
		names[f.Name] = f.Kind
	}

	if _, ok := names["createUser"]; !ok {
		t.Error("expected createUser to be detected")
	}
	if _, ok := names["buildOrder"]; !ok {
		t.Error("expected buildOrder to be detected")
	}
	if k := names["setupTestEnv"]; k != models.FixtureSetupHook {
		t.Errorf("setupTestEnv: want setup_hook, got %s", k)
	}
	if k := names["mockAuthService"]; k != models.FixtureMockProvider {
		t.Errorf("mockAuthService: want mock_provider, got %s", k)
	}
}

func TestDetectJSFixtures_TestData(t *testing.T) {
	t.Parallel()
	src := `
const testData = [{ id: 1 }, { id: 2 }];
const mockData = { users: [] };
const sampleData = {};
`
	fixtures := detectJSFixtures(src, "test/data.ts", "jest")

	names := map[string]bool{}
	for _, f := range fixtures {
		names[f.Name] = true
		if f.Kind != models.FixtureDataLoader {
			t.Errorf("%s: want kind data_loader, got %s", f.Name, f.Kind)
		}
	}

	for _, name := range []string{"testData", "mockData", "sampleData"} {
		if !names[name] {
			t.Errorf("expected %q to be detected as test data", name)
		}
	}
}

func TestDetectJSFixtures_SharedFile(t *testing.T) {
	t.Parallel()
	src := `
beforeEach(() => {
  globalSetup();
});
`
	// File in a helpers directory should be marked as shared.
	fixtures := detectJSFixtures(src, "test/helpers/setup.ts", "jest")

	if len(fixtures) == 0 {
		t.Fatal("expected at least one fixture")
	}
	for _, f := range fixtures {
		if !f.Shared {
			t.Errorf("fixture %q in helpers/ should be shared", f.Name)
		}
	}
}

func TestDetectPythonFixtures_PytestFixture(t *testing.T) {
	t.Parallel()
	src := `
import pytest

@pytest.fixture
def db_session():
    session = create_session()
    yield session
    session.close()

@pytest.fixture(scope="session")
def app_client():
    return create_app().test_client()

@pytest.fixture(scope="module")
def seed_data():
    return load_seed()
`
	fixtures := detectPythonFixtures(src, "conftest.py", "pytest")

	names := map[string]string{}
	for _, f := range fixtures {
		names[f.Name] = f.Scope
	}

	if scope, ok := names["db_session"]; !ok {
		t.Error("expected db_session fixture")
	} else if scope != "test" {
		t.Errorf("db_session: want scope test, got %s", scope)
	}

	if scope, ok := names["app_client"]; !ok {
		t.Error("expected app_client fixture")
	} else if scope != "session" {
		t.Errorf("app_client: want scope session, got %s", scope)
	}

	if scope, ok := names["seed_data"]; !ok {
		t.Error("expected seed_data fixture")
	} else if scope != "module" {
		t.Errorf("seed_data: want scope module, got %s", scope)
	}

	// conftest.py fixtures should be shared.
	for _, f := range fixtures {
		if !f.Shared {
			t.Errorf("conftest.py fixture %q should be shared", f.Name)
		}
	}
}

func TestDetectPythonFixtures_SetUpTearDown(t *testing.T) {
	t.Parallel()
	src := `
class TestUser(unittest.TestCase):
    def setUp(self):
        self.db = connect()

    def tearDown(self):
        self.db.close()

    def setUpClass(cls):
        cls.server = start_server()
`
	fixtures := detectPythonFixtures(src, "test_user.py", "unittest")

	kinds := map[string]models.FixtureKind{}
	scopes := map[string]string{}
	for _, f := range fixtures {
		kinds[f.Name] = f.Kind
		scopes[f.Name] = f.Scope
	}

	if kinds["setUp"] != models.FixtureSetupHook {
		t.Errorf("setUp: want setup_hook, got %s", kinds["setUp"])
	}
	if kinds["tearDown"] != models.FixtureTeardownHook {
		t.Errorf("tearDown: want teardown_hook, got %s", kinds["tearDown"])
	}
	if scopes["setUpClass"] != "suite" {
		t.Errorf("setUpClass: want scope suite, got %s", scopes["setUpClass"])
	}
}

func TestDetectPythonFixtures_NestedFixtures(t *testing.T) {
	t.Parallel()
	src := `
import pytest

@pytest.fixture
def base_user():
    return {"name": "test"}

@pytest.fixture
def admin_user(base_user):
    return {**base_user, "role": "admin"}
`
	fixtures := detectPythonFixtures(src, "conftest.py", "pytest")

	names := map[string]bool{}
	for _, f := range fixtures {
		names[f.Name] = true
	}

	if !names["base_user"] {
		t.Error("expected base_user fixture")
	}
	if !names["admin_user"] {
		t.Error("expected admin_user (nested fixture)")
	}
}

func TestDetectGoFixtures_TestHelpers(t *testing.T) {
	t.Parallel()
	src := `package auth_test

import "testing"

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler())
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	// ...
}
`
	fixtures := detectGoFixtures(src, "internal/auth/auth_test.go")

	names := map[string]models.FixtureKind{}
	for _, f := range fixtures {
		names[f.Name] = f.Kind
	}

	if _, ok := names["setupTestDB"]; !ok {
		t.Error("expected setupTestDB to be detected")
	}
	if _, ok := names["newTestServer"]; !ok {
		t.Error("expected newTestServer to be detected")
	}

	// Verify detection tier — Go test helpers accepting *testing.T are structural.
	for _, f := range fixtures {
		if f.Name == "setupTestDB" || f.Name == "newTestServer" {
			if f.DetectionTier != models.TierStructural {
				t.Errorf("%s: want tier structural, got %s", f.Name, f.DetectionTier)
			}
		}
	}
}

func TestDetectGoFixtures_TestMain(t *testing.T) {
	t.Parallel()
	src := `package main_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}
`
	fixtures := detectGoFixtures(src, "main_test.go")

	found := false
	for _, f := range fixtures {
		if f.Name == "TestMain" {
			found = true
			if f.Kind != models.FixtureSetupHook {
				t.Errorf("TestMain: want kind setup_hook, got %s", f.Kind)
			}
			if f.Scope != "module" {
				t.Errorf("TestMain: want scope module, got %s", f.Scope)
			}
			if f.Confidence != 0.99 {
				t.Errorf("TestMain: want confidence 0.99, got %f", f.Confidence)
			}
		}
	}
	if !found {
		t.Error("expected TestMain fixture to be detected")
	}
}

func TestDetectGoFixtures_MockBuilders(t *testing.T) {
	t.Parallel()
	src := `package service_test

func mockEmailService() *EmailService {
	return &EmailService{send: func(to, body string) error { return nil }}
}

func fakeDatabase() *DB {
	return &DB{memory: true}
}

var testUsers = []User{
	{ID: 1, Name: "Alice"},
	{ID: 2, Name: "Bob"},
}
`
	fixtures := detectGoFixtures(src, "service_test.go")

	kinds := map[string]models.FixtureKind{}
	for _, f := range fixtures {
		kinds[f.Name] = f.Kind
	}

	if kinds["mockEmailService"] != models.FixtureMockProvider {
		t.Errorf("mockEmailService: want mock_provider, got %s", kinds["mockEmailService"])
	}
	if kinds["fakeDatabase"] != models.FixtureMockProvider {
		t.Errorf("fakeDatabase: want mock_provider, got %s", kinds["fakeDatabase"])
	}
	if kinds["testUsers"] != models.FixtureDataLoader {
		t.Errorf("testUsers: want data_loader, got %s", kinds["testUsers"])
	}
}

func TestDetectJavaFixtures_Lifecycle(t *testing.T) {
	t.Parallel()
	src := `
import org.junit.jupiter.api.*;

class UserServiceTest {

    @BeforeEach
    public void setUp() {
        db = new TestDatabase();
    }

    @AfterEach
    public void tearDown() {
        db.close();
    }

    @BeforeAll
    static void initAll() {
        server = startServer();
    }

    @Test
    void testFindUser() {
        // ...
    }
}
`
	fixtures := detectJavaFixtures(src, "src/test/UserServiceTest.java")

	kinds := map[string]models.FixtureKind{}
	scopes := map[string]string{}
	for _, f := range fixtures {
		kinds[f.Name] = f.Kind
		scopes[f.Name] = f.Scope
	}

	if kinds["setUp"] != models.FixtureSetupHook {
		t.Errorf("setUp: want setup_hook, got %s", kinds["setUp"])
	}
	if kinds["tearDown"] != models.FixtureTeardownHook {
		t.Errorf("tearDown: want teardown_hook, got %s", kinds["tearDown"])
	}
	if scopes["initAll"] != "suite" {
		t.Errorf("initAll: want scope suite, got %s", scopes["initAll"])
	}
}

func TestIsSharedFixtureFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path string
		want bool
	}{
		{"conftest.py", true},
		{"tests/conftest.py", true},
		{"test/helpers.ts", true},
		{"test/factories.ts", true},
		{"test/__fixtures__/data.ts", true},
		{"test/support/setup.ts", true},
		{"test/test-utils.ts", true},
		{"test/testutils.ts", true},
		{"test/mocks.ts", true},
		{"test/UserComponent.test.ts", false},
		{"src/auth/login.ts", false},
	}

	for _, tc := range cases {
		if got := isSharedFixtureFile(tc.path); got != tc.want {
			t.Errorf("isSharedFixtureFile(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestExtractFixtures_Integration(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeTempFile(t, root, "test/helpers.ts", `
export function createUser(data = {}) {
  return { id: 1, ...data };
}

const mockApiClient = jest.fn();
`)
	writeTempFile(t, root, "test/app.test.ts", `
import { createUser } from './helpers';

describe('App', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('works', () => {
    const user = createUser();
    expect(user.id).toBe(1);
  });
});
`)

	testFiles := []models.TestFile{
		{Path: "test/helpers.ts", Framework: "jest"},
		{Path: "test/app.test.ts", Framework: "jest"},
	}

	fixtures := ExtractFixtures(root, testFiles)

	if len(fixtures) < 2 {
		t.Fatalf("expected at least 2 fixtures, got %d", len(fixtures))
	}

	// Verify we got fixtures from both files.
	paths := map[string]bool{}
	for _, f := range fixtures {
		paths[f.Path] = true
	}
	if !paths["test/helpers.ts"] {
		t.Error("expected fixtures from test/helpers.ts")
	}
	if !paths["test/app.test.ts"] {
		t.Error("expected fixtures from test/app.test.ts")
	}
}

func TestFixtureReuse_AcrossFiles(t *testing.T) {
	t.Parallel()
	// Test that shared fixtures from helpers files are marked as shared.
	helperSrc := `
export function createUser(overrides = {}) {
  return { id: 1, name: 'test', ...overrides };
}

export function buildOrder(items) {
  return { items, total: items.length };
}
`
	fixtures := detectJSFixtures(helperSrc, "test/helpers/factories.ts", "jest")

	for _, f := range fixtures {
		if !f.Shared {
			t.Errorf("fixture %q from factories.ts should be shared", f.Name)
		}
	}
}
