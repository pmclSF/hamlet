import pytest

@pytest.fixture(scope="session")
def db_connection():
    return "connected"

def test_db(db_connection):
    assert db_connection == "connected"
