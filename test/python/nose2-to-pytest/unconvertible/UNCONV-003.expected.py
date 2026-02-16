# HAMLET-TODO [UNCONVERTIBLE-SUCH-DSL]: nose2 such DSL has no direct pytest equivalent

# Original: it = such.A('calculator')

# Manual action required: Rewrite using standard pytest test functions or classes

it = such.A('calculator')

@it.should('add numbers')
def test_add(case):
    assert 1 + 1 == 2
