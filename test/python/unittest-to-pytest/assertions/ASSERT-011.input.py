import unittest

class TestAssert(unittest.TestCase):
    def test_not_almost_equal(self):
        self.assertNotAlmostEqual(1.0, 2.0)
