import pytest


@pytest.mark.parametrize("params", [1, 2, 3])
def test_param(value):
    assert value > 0
