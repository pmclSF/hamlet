import unittest


class TestSkipped(unittest.TestCase):
    @unittest.skip
    def test_skipped(self):
        self.assertTrue(True)
