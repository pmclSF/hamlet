import unittest


class TestNone(unittest.TestCase):
    def test_none(self):
        result = None
        self.assertIsNone(result)
