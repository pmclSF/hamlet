import unittest

class TestSkip(unittest.TestCase):
    @unittest.skip("not ready")
    def test_skipped(self):
        self.assertTrue(False)
