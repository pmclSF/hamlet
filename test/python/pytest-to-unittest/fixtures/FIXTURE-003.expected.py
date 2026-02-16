import pytest
import unittest


class TestExample(unittest.TestCase):
    # HAMLET-TODO [UNCONVERTIBLE-FIXTURE]: pytest fixture without autouse=True has no direct unittest equivalent
    # Original: @pytest.fixture
    # Manual action required: Manually convert this fixture to setUp/tearDown or pass the value directly
    @pytest.fixture
    def my_data(self):
        return 42

    def test_example(self, my_data):
        self.assertEqual(my_data, 42)
