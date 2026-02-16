import sys
import unittest


class TestUnix(unittest.TestCase):
    @unittest.skipIf(sys.platform == "win32", "not on windows")
    def test_unix(self):
        self.assertTrue(True)
