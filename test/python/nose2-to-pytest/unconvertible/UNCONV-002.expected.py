# HAMLET-TODO [UNCONVERTIBLE-NOSE-PLUGIN]: nose2 plugin has no direct pytest equivalent
# Original: from nose2.tools.decorators import with_setup
# Manual action required: Find a pytest plugin or built-in feature that provides equivalent functionality
from nose2.tools.decorators import with_setup

def setup_func():
    pass

@with_setup(setup_func)
def test_with_setup():
    assert True
