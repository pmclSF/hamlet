import unittest

class TestAssert(unittest.TestCase):
    def test_warns(self):
        with self.assertWarns(DeprecationWarning):
            some_deprecated_func()
