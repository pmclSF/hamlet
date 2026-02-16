import unittest


class TestNotNone(unittest.TestCase):
    def test_not_none(self):
        result = 42
        self.assertIsNotNone(result)
