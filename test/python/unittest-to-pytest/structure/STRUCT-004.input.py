import unittest

class TestMultiple(unittest.TestCase):
    def test_first(self):
        self.assertTrue(True)

    def test_second(self):
        self.assertFalse(False)

    def test_third(self):
        self.assertEqual(1, 1)
