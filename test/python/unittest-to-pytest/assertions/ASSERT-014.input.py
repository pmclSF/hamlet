import unittest

class TestAssert(unittest.TestCase):
    def test_regex(self):
        self.assertRegex("hello world", r"hello\s\w+")
