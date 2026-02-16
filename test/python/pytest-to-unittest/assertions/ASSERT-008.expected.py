import unittest


class TestNotIn(unittest.TestCase):
    def test_not_in(self):
        self.assertNotIn(4, [1, 2, 3])
