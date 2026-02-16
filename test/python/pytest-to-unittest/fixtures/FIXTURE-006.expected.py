import pytest
import unittest


class TestExample(unittest.TestCase):
    def setUp(self):
        print("setup")

    # HAMLET-TODO [UNCONVERTIBLE-FIXTURE]: pytest fixture without autouse=True has no direct unittest equivalent
    # Original: @pytest.fixture
    # Manual action required: Manually convert this fixture to setUp/tearDown or pass the value directly
    @pytest.fixture
    def data(self):
        return [1, 2, 3]

    def test_example(self):
        self.assertTrue(True)
