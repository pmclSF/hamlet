import unittest

class TestErrors(unittest.TestCase):
    def test_raises(self):
        with self.assertRaises(ValueError):
            int("abc")
