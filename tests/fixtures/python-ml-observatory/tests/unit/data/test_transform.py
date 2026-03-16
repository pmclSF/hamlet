"""Tests for data transforms."""
from src.data.transform import normalize_text, tokenize, encode_labels

def test_normalize_text():
    assert normalize_text("  Hello World  ") == "hello world"

def test_tokenize():
    assert tokenize("hello world") == ["hello", "world"]

def test_encode_labels():
    encoded, mapping = encode_labels(["pos", "neg", "pos"])
    assert len(encoded) == 3
    assert mapping["pos"] == 0
