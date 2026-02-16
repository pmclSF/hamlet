import pytest
import unittest


class TestSlow(unittest.TestCase):
    @pytest.mark.slow
    def test_slow(self):
        self.assertTrue(True)
