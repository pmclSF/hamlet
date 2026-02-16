import unittest

class TestSelf(unittest.TestCase):
    def test_no_self(self):
        result = 42
        self.assertEqual(result, 42)
