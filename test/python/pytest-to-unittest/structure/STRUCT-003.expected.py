import os
import unittest


class TestPath(unittest.TestCase):
    def test_path(self):
        self.assertTrue(os.path.exists("/"))
