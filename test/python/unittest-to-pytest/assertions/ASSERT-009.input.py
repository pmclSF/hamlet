import unittest

class TestAssert(unittest.TestCase):
    def test_instance(self):
        self.assertIsInstance(42, int)
