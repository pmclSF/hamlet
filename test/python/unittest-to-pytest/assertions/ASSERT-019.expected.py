import pytest

def test_raises_inline():
    with pytest.raises(ValueError):
        int("abc")
