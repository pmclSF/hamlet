import unittest

class TestIndent(unittest.TestCase):
    def test_indented(self):
        x = 1
        y = 2
        self.assertEqual(x + y, 3)
