import pytest


@pytest.mark.parametrize("params", [1, 2, 3])
def test_values(val):
    assert val == val
