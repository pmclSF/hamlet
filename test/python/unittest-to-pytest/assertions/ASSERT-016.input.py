import unittest

class TestAssert(unittest.TestCase):
    def test_multiline_equal(self):
        self.assertMultiLineEqual("line1\nline2", "line1\nline2")
