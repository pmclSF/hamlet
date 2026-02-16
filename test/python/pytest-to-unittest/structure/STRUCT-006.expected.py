import unittest


class TestValue(unittest.TestCase):
    def test_value(self):
        x = 42
        self.assertEqual(x, 42)
