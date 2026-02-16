import pytest

def test_capsys(capsys):
    print("output")
    captured = capsys.readouterr()
    assert captured.out == "output\n"
