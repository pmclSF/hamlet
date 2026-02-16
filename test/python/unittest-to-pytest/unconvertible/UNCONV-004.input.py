import unittest

def tearDownModule():
    print("module teardown")

class TestModule(unittest.TestCase):
    def test_example(self):
        self.assertTrue(True)
