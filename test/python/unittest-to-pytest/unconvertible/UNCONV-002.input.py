import unittest

class TestCleanup(unittest.TestCase):
    def test_cleanup(self):
        self.addCleanup(print, "cleanup")
        self.assertTrue(True)
