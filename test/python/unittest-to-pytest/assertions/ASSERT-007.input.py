import unittest

class TestAssert(unittest.TestCase):
    def test_in(self):
        self.assertIn(1, [1, 2, 3])
