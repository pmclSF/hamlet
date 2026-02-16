import unittest

class TestNoFixtures(unittest.TestCase):
    def test_simple(self):
        self.assertEqual(1, 1)
