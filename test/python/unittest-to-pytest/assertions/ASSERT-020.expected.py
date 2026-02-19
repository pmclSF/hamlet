import pytest

def test_raises_regex_inline():
    with pytest.raises(ValueError, match="invalid literal"):
        int("abc")
