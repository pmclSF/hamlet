import unittest

class TestSubTestEqual(unittest.TestCase):
    def test_values(self):
        cases = [(1, 1), (2, 2), (3, 3)]
        for a, b in cases:
            with self.subTest(a=a, b=b):
                self.assertEqual(a, b)
