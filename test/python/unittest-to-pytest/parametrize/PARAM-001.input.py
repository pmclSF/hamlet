import unittest

class TestSubTest(unittest.TestCase):
    def test_subtests(self):
        for i in range(3):
            with self.subTest(i=i):
                self.assertTrue(i >= 0)
