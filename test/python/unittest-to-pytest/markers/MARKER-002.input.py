import unittest
import sys

class TestSkipIf(unittest.TestCase):
    @unittest.skipIf(sys.platform == "win32", "not on windows")
    def test_unix_only(self):
        self.assertTrue(True)
