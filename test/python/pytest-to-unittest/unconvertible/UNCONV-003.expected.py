import unittest


class TestFile(unittest.TestCase):
    # TERRAIN-TODO [UNCONVERTIBLE-TMPPATH]: tmp_path fixture has no direct unittest equivalent
    # Original: def test_file(self, tmp_path):
    # Manual action required: Use tempfile.mkdtemp() in setUp/tearDown
    def test_file(self, tmp_path):
        # TERRAIN-TODO [UNCONVERTIBLE-TMPPATH]: tmp_path fixture has no direct unittest equivalent
        # Original: f = tmp_path / "test.txt"
        # Manual action required: Use tempfile.mkdtemp() in setUp/tearDown
        f = tmp_path / "test.txt"
        f.write_text("hello")
        self.assertEqual(f.read_text(), "hello")
