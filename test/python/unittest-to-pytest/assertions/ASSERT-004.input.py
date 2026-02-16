import unittest

class TestAssert(unittest.TestCase):
    def test_false(self):
        self.assertFalse(False)
