import unittest

class TestNested(unittest.TestCase):
    def test_loop(self):
        total = 0
        for i in range(5):
            total += i
        self.assertEqual(total, 10)
