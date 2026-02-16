import unittest

class TestAssert(unittest.TestCase):
    def test_not_none(self):
        self.assertIsNotNone(42)
