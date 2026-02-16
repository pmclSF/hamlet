import pytest

def test_raises():
    with pytest.raises(ValueError):
        int("abc")
