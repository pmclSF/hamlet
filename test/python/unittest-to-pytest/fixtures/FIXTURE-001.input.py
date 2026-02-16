import unittest

class TestSetUp(unittest.TestCase):
    def setUp(self):
        print("setting up")

    def test_example(self):
        self.assertTrue(True)
