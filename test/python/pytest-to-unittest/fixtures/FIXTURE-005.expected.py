import pytest
import unittest


class TestPositive(unittest.TestCase):
    # HAMLET-TODO [UNCONVERTIBLE-FIXTURE]: pytest fixture without autouse=True has no direct unittest equivalent
    # Original: @pytest.fixture(params=[1, 2, 3])
    # Manual action required: Manually convert this fixture to setUp/tearDown or pass the value directly
    @pytest.fixture(params=[1, 2, 3])
    def number(self, request):
        return request.param

    def test_positive(self, number):
        self.assertGreater(number, 0)
