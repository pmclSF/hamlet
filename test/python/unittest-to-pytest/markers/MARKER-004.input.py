import unittest

class TestExpected(unittest.TestCase):
    @unittest.expectedFailure
    def test_expected_fail(self):
        self.assertEqual(1, 2)
