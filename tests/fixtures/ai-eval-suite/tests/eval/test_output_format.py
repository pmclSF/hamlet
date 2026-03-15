"""Test model output format compliance."""
from src.models.classifier import classify, CATEGORIES


class TestOutputFormat:
    def test_classify_returns_tuple(self):
        result = classify("test input")
        assert isinstance(result, tuple)
        assert len(result) == 2

    def test_category_is_valid(self):
        category, _ = classify("some text")
        assert category in CATEGORIES

    def test_confidence_in_range(self):
        _, confidence = classify("some text")
        assert 0.0 <= confidence <= 1.0

    def test_confidence_type(self):
        _, confidence = classify("test")
        assert isinstance(confidence, float)
