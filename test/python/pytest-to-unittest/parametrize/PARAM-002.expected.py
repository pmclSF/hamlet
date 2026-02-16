import pytest
import unittest


class TestPositive(unittest.TestCase):
    # HAMLET-TODO [UNCONVERTIBLE-PARAMETRIZE]: @pytest.mark.parametrize has no direct unittest equivalent
    # Original: @pytest.mark.parametrize("x", [1, 2, 3])
    # Manual action required: Use subTest() or create individual test methods for each parameter set
    @pytest.mark.parametrize("x", [1, 2, 3])
    def test_positive(self, x):
        self.assertGreater(x, 0)

    # HAMLET-TODO [UNCONVERTIBLE-PARAMETRIZE]: @pytest.mark.parametrize has no direct unittest equivalent

    # Original: @pytest.mark.parametrize("s", ["hello", "world"])

    # Manual action required: Use subTest() or create individual test methods for each parameter set

    @pytest.mark.parametrize("s", ["hello", "world"])
    def test_string(self, s):
        self.assertGreater(len(s), 0)
