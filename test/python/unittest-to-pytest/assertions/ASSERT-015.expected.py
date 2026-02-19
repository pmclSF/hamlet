import re

def test_not_regex():
    assert not re.search(r"^goodbye", "hello world")
