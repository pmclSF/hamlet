import unittest

class TestBoth(unittest.TestCase):
    def setUp(self):
        print("setting up")

    def tearDown(self):
        print("cleaning up")

    def test_example(self):
        self.assertTrue(True)
