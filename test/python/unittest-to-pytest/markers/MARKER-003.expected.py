import pytest
import sys

@pytest.mark.skipif(not sys.platform == "linux", reason="linux only")
def test_linux_only():
    assert True
