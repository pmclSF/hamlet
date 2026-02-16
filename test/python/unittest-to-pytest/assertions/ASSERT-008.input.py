import unittest

class TestAssert(unittest.TestCase):
    def test_not_in(self):
        self.assertNotIn(4, [1, 2, 3])
