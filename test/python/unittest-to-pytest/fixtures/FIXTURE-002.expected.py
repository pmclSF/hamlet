import pytest

@pytest.fixture(autouse=True)
def teardown():
    yield
    print("cleaning up")

def test_example():
    assert True
