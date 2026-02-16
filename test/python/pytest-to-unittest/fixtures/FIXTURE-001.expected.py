import unittest


class TestExample(unittest.TestCase):
    def setUp(self):
        print("setting up")

    def test_example(self):
        self.assertTrue(True)
