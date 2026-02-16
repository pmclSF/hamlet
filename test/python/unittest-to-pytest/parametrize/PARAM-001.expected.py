def test_subtests():
    for i in range(3):
        with self.subTest(i=i):
            assert i >= 0
