import nose2
from nose2.tools.decorators import with_setup

def setup_func():
    pass

@with_setup(setup_func)
def test_with_setup():
    assert True
