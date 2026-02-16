import unittest


class TestSkipped(unittest.TestCase):
    @unittest.skip("not ready")
    def test_skipped(self):
        self.assertTrue(True)
