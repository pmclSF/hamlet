import unittest
import sys

class TestSkipUnless(unittest.TestCase):
    @unittest.skipUnless(sys.platform == "linux", "linux only")
    def test_linux_only(self):
        self.assertTrue(True)
