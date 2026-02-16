import pytest

@pytest.fixture(autouse=True)
def setup_teardown():
    print("setting up")
    yield
    print("tearing down")

def test_example():
    assert True
