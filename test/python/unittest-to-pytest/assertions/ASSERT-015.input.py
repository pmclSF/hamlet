import unittest

class TestAssert(unittest.TestCase):
    def test_not_regex(self):
        self.assertNotRegex("hello world", r"^goodbye")
