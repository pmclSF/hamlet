# HAMLET-TODO [UNCONVERTIBLE-MODULE-SETUP]: Module-level setup/teardown has no direct pytest equivalent in-file
# Original: def setUpModule():
# Manual action required: Move to conftest.py as a session/module-scoped fixture
def setUpModule():
    print("module setup")

def test_example():
    assert True
