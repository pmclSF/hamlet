import pytest


def test_equal():
    assert 1 == 1

def test_error():
    with pytest.raises(ValueError):
