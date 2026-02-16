import pytest

@pytest.fixture
def my_data():
    return 42

def test_example(my_data):
    assert my_data == 42
