import unittest

def setUpModule():
    print("module setup")

class TestModule(unittest.TestCase):
    def test_example(self):
        self.assertTrue(True)
