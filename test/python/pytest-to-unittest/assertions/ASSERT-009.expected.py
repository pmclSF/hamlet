import unittest


class TestInstance(unittest.TestCase):
    def test_instance(self):
        self.assertIsInstance(42, int)
