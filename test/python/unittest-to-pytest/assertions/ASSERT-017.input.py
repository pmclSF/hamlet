import unittest

class TestAssert(unittest.TestCase):
    def test_sequence_equal(self):
        self.assertSequenceEqual((1, 2, 3), [1, 2, 3])
