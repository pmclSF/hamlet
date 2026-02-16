import pytest
import unittest


class TestIncrement(unittest.TestCase):
    # HAMLET-TODO [UNCONVERTIBLE-PARAMETRIZE]: @pytest.mark.parametrize has no direct unittest equivalent
    # Original: @pytest.mark.parametrize("x,expected", [(1, 2), (2, 3)])
    # Manual action required: Use subTest() or create individual test methods for each parameter set
    @pytest.mark.parametrize("x,expected", [(1, 2), (2, 3)])
    def test_increment(self, x, expected):
        self.assertEqual(x + 1, expected)
