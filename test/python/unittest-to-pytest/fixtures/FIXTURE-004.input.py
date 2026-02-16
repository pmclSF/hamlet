import unittest

class TestSetUpMulti(unittest.TestCase):
    def setUp(self):
        print("step one")
        print("step two")

    def test_example(self):
        self.assertTrue(True)
