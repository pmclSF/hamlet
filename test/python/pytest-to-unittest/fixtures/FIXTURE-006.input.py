import pytest

@pytest.fixture(autouse=True)
def setup():
    print("setup")

@pytest.fixture
def data():
    return [1, 2, 3]

def test_example():
    assert True
