import pytest

@pytest.mark.skip(reason="not ready")
def test_skipped():
    assert False
