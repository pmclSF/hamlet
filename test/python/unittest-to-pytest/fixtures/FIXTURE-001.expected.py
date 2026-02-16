import pytest

@pytest.fixture(autouse=True)
def setup():
    print("setting up")

def test_example():
    assert True
