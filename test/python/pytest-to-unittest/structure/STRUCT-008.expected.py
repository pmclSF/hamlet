import unittest


class TestMultiLine(unittest.TestCase):
    def test_multi_line(self):
        a = 1
        b = 2
        c = a + b
        self.assertEqual(c, 3)
