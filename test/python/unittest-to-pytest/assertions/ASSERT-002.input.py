import unittest

class TestAssert(unittest.TestCase):
    def test_not_equal(self):
        self.assertNotEqual(1, 2)
