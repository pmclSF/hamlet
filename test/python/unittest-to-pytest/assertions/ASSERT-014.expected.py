import re

def test_regex():
    assert re.search(r"hello\s\w+", "hello world")
