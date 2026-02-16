import unittest

class TestLoopAssert(unittest.TestCase):
    def test_loop_without_subtest(self):
        for x in [1, 2, 3]:
            self.assertGreater(x, 0)
