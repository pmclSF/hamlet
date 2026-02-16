import pytest
import sys

@pytest.mark.skipif(sys.platform == "win32", reason="not on windows")
def test_unix():
    assert True
