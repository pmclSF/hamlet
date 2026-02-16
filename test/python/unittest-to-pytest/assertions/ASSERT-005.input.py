import unittest

class TestAssert(unittest.TestCase):
    def test_none(self):
        self.assertIsNone(None)
