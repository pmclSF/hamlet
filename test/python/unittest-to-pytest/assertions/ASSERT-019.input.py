import unittest

class TestAssert(unittest.TestCase):
    def test_raises_inline(self):
        self.assertRaises(ValueError, int, "abc")
