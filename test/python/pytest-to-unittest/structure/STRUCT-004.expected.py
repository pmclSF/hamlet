import unittest


class TestNested(unittest.TestCase):
    def test_nested(self):
        for i in range(3):
            self.assertGreaterEqual(i, 0)
