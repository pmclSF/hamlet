import pytest

@pytest.fixture(autouse=True)
def setup_data():
    print("setting up")

def test_example():
    assert True
