def test_generator():
    for i in range(3):
        yield check_value, i, i

def check_value(a, b):
    assert a == b
