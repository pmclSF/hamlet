import pytest

@pytest.fixture(autouse=True)
def setup():
    print("step one")
    print("step two")

def test_example():
    assert True
