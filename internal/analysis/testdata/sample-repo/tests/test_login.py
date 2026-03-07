import pytest

def test_login_success():
    assert login("admin", "pass") is True
