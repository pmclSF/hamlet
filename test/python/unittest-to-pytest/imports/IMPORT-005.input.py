import unittest
import os
from collections import OrderedDict

class TestExample(unittest.TestCase):
    def test_os(self):
        self.assertTrue(os.path.exists("/"))
