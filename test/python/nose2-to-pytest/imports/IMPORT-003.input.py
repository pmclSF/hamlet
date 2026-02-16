from nose.tools import assert_equal, assert_raises

def test_equal():
    assert_equal(1, 1)

def test_error():
    assert_raises(ValueError)
