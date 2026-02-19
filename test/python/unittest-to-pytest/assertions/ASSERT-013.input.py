import unittest

class TestAssert(unittest.TestCase):
    def test_count_equal(self):
        self.assertCountEqual([3, 1, 2], [1, 2, 3])
