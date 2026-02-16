import pytest

@pytest.mark.parametrize("x", [1, 2, 3])
def test_positive(x):
    assert x > 0

@pytest.mark.parametrize("s", ["hello", "world"])
def test_string(s):
    assert len(s) > 0
