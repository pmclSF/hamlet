import pytest

@pytest.mark.parametrize("x", [1, 2])
@pytest.mark.parametrize("y", [3, 4])
def test_combo(x, y):
    assert x + y > 0
