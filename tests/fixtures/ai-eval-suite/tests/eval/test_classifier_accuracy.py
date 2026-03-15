"""Test classifier accuracy on evaluation dataset."""
import pytest
from src.models.classifier import classify, batch_classify
from src.eval.dataset import load_eval_dataset
from src.eval.metrics import accuracy


class TestClassifierAccuracy:
    def test_accuracy_above_threshold(self):
        """Model accuracy should exceed 60% on eval dataset."""
        texts, labels = load_eval_dataset()
        predictions = [classify(t)[0] for t in texts]
        acc = accuracy(predictions, labels)
        assert acc >= 0.6, f"Accuracy {acc:.2f} below threshold 0.60"

    def test_positive_classification(self):
        category, confidence = classify("This is amazing and great")
        assert category == "positive"
        assert confidence > 0.5

    def test_negative_classification(self):
        category, confidence = classify("This is terrible and awful")
        assert category == "negative"
        assert confidence > 0.5

    def test_neutral_classification(self):
        category, confidence = classify("The package arrived today")
        assert category == "neutral"

    def test_batch_classification(self):
        results = batch_classify(["great product", "awful service", "it works"])
        assert len(results) == 3
        assert results[0][0] == "positive"
        assert results[1][0] == "negative"
