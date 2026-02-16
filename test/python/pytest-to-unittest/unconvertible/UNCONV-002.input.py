import pytest

def test_output(capfd):
    print("hello")
    captured = capfd.readouterr()
    assert captured.out == "hello\n"
