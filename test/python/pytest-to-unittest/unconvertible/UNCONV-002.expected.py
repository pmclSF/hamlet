import unittest


class TestOutput(unittest.TestCase):
    # HAMLET-TODO [UNCONVERTIBLE-CAPTURE]: pytest capture fixtures have no direct unittest equivalent
    # Original: def test_output(self, capfd):
    # Manual action required: Use contextlib.redirect_stdout or unittest.mock to capture output
    def test_output(self, capfd):
        print("hello")
        # HAMLET-TODO [UNCONVERTIBLE-CAPTURE]: pytest capture fixtures have no direct unittest equivalent
        # Original: captured = capfd.readouterr()
        # Manual action required: Use contextlib.redirect_stdout or unittest.mock to capture output
        captured = capfd.readouterr()
        self.assertEqual(captured.out, "hello\n")
