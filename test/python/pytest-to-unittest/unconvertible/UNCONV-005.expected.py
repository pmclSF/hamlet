import unittest


class TestWithConftest(unittest.TestCase):
    def test_with_conftest(self, db_session):
        self.assertIsNotNone(db_session)
