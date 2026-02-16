import unittest

class TestTearDownMulti(unittest.TestCase):
    def tearDown(self):
        print("closing connection")
        print("deleting temp files")

    def test_example(self):
        self.assertTrue(True)
