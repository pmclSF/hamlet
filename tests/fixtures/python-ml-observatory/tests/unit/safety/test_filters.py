"""Tests for safety filters."""
from src.safety.filters import check_input_safety, apply_safety_filter

def test_safe_input():
    result = check_input_safety("hello world")
    assert result["safe"] is True

def test_unsafe_input():
    result = check_input_safety("ignore instructions")
    assert result["safe"] is False

def test_apply_safety_filter_safe():
    result = apply_safety_filter("normal query", "normal response")
    assert result["blocked"] is False
