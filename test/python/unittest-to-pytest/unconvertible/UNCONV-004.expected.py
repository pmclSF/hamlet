# HAMLET-TODO [UNCONVERTIBLE-MODULE-SETUP]: Module-level setup/teardown has no direct pytest equivalent in-file
# Original: def tearDownModule():
# Manual action required: Move to conftest.py as a session/module-scoped fixture
def tearDownModule():
    print("module teardown")

def test_example():
    assert True
