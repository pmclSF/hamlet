def test_values():
    cases = [(1, 1), (2, 2), (3, 3)]
    for a, b in cases:
        with self.subTest(a=a, b=b):
            assert a == b
