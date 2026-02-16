import pytest

@pytest.fixture(autouse=True)
def teardown():
    yield
    print("closing connection")
    print("deleting temp files")

def test_example():
    assert True
