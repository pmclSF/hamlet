import pytest

@pytest.mark.parametrize("x,expected", [(1, 2), (2, 3)])
def test_increment(x, expected):
    assert x + 1 == expected
