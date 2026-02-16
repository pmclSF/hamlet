import pytest

def test_monkeypatch(monkeypatch):
    monkeypatch.setattr("os.getcwd", lambda: "/fake")
    assert True
