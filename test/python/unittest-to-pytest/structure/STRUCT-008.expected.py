def test_loop():
    total = 0
    for i in range(5):
        total += i
    assert total == 10
