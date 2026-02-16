import unittest


class TestCapsys(unittest.TestCase):
    # HAMLET-TODO [UNCONVERTIBLE-CAPTURE]: pytest capture fixtures have no direct unittest equivalent
    # Original: def test_capsys(self, capsys):
    # Manual action required: Use contextlib.redirect_stdout or unittest.mock to capture output
    def test_capsys(self, capsys):
        print("output")
        # HAMLET-TODO [UNCONVERTIBLE-CAPTURE]: pytest capture fixtures have no direct unittest equivalent
        # Original: captured = capsys.readouterr()
        # Manual action required: Use contextlib.redirect_stdout or unittest.mock to capture output
        captured = capsys.readouterr()
        self.assertEqual(captured.out, "output\n")
