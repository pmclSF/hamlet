import pytest

def test_not_almost_equal():
    assert 1.0 != pytest.approx(2.0)
