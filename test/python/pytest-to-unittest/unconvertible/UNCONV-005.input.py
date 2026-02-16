import pytest

def test_with_conftest(db_session):
    assert db_session is not None
