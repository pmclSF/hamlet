import unittest

class TestParams(unittest.TestCase):
    def test_with_param(self):
        value = 42
        self.assertEqual(value, 42)
