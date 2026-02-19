import unittest

class TestAssert(unittest.TestCase):
    def test_raises_regex_inline(self):
        self.assertRaisesRegex(ValueError, "invalid literal", int, "abc")
