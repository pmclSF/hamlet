import pytest
import unittest


class TestDb(unittest.TestCase):
    # HAMLET-TODO [UNCONVERTIBLE-FIXTURE]: pytest fixture without autouse=True has no direct unittest equivalent
    # Original: @pytest.fixture(scope="session")
    # Manual action required: Manually convert this fixture to setUp/tearDown or pass the value directly
    @pytest.fixture(scope="session")
    def db_connection(self):
        return "connected"

    def test_db(self, db_connection):
        self.assertEqual(db_connection, "connected")
