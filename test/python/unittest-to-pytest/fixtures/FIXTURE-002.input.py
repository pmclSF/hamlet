import unittest

class TestTearDown(unittest.TestCase):
    def tearDown(self):
        print("cleaning up")

    def test_example(self):
        self.assertTrue(True)
