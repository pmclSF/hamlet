import pytest

def test_warns():
    with pytest.warns(DeprecationWarning):
        some_deprecated_func()
