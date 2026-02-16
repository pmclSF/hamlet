import pytest
import sys

@pytest.mark.skipif(sys.platform == "win32", reason="not on windows")
def test_unix_only():
    assert True
