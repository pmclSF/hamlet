from nose.tools import assert_equal
from nose2.tools import params

@params(1, 2, 3)
def test_values(val):
    assert_equal(val, val)
