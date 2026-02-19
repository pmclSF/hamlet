import unittest

class TestAssert(unittest.TestCase):
    def test_not_is_instance(self):
        self.assertNotIsInstance("hello", int)
