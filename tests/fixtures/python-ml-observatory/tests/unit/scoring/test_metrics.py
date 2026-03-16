"""Tests for metrics."""
from src.scoring.metrics import accuracy, precision_recall

def test_accuracy():
    assert accuracy(["pos", "neg", "pos"], ["pos", "neg", "neg"]) > 0.6

def test_precision_recall():
    prec, rec = precision_recall(["pos", "neg"], ["pos", "pos"], "pos")
    assert prec == 1.0
    assert rec == 0.5
