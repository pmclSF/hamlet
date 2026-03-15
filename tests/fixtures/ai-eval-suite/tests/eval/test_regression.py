"""Regression tests for known-good classifications."""
import pytest
from src.models.classifier import classify


class TestRegressions:
    @pytest.mark.parametrize("text,expected", [
        ("I love this product", "positive"),
        ("This is the worst", "negative"),
        ("It arrived on Tuesday", "neutral"),
    ])
    def test_known_classifications(self, text, expected):
        category, _ = classify(text)
        assert category == expected, f"Regression: '{text}' classified as {category}, expected {expected}"

    def test_empty_string_does_not_crash(self):
        category, confidence = classify("")
        assert category in ["positive", "negative", "neutral"]
        assert confidence >= 0

    @pytest.mark.skip(reason="Model v2 not yet deployed")
    def test_sarcasm_detection(self):
        category, _ = classify("Oh sure, this is just GREAT")
        assert category == "negative"

    @pytest.mark.skip(reason="Multilingual support pending")
    def test_french_classification(self):
        category, _ = classify("C'est magnifique")
        assert category == "positive"
