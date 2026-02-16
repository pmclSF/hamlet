from nose2.tools import params
from nose.tools import assert_true

@params(1, 2, 3)
def test_param(value):
    assert_true(value > 0)
