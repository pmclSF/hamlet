import unittest

class TestAssert(unittest.TestCase):
    def test_raises(self):
        with self.assertRaises(ValueError):
            int("abc")
