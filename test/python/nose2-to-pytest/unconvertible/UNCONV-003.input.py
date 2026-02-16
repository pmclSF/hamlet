from nose.tools import assert_equal

it = such.A('calculator')

@it.should('add numbers')
def test_add(case):
    assert_equal(1 + 1, 2)
