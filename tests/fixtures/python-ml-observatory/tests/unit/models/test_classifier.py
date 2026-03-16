"""Tests for classifier."""
from src.models.classifier import classify, batch_classify

def test_classify_positive():
    label, conf = classify("great product")
    assert label == "positive"
    assert conf > 0.5

def test_classify_negative():
    label, conf = classify("terrible service")
    assert label == "negative"

def test_batch_classify():
    results = batch_classify(["good", "bad"])
    assert len(results) == 2
