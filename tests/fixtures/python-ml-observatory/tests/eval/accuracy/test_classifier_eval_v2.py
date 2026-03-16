"""Classifier accuracy evaluation v2 (duplicate)."""
from src.models.classifier import classify, batch_classify
from src.data.loader import load_eval_dataset
from src.conftest_helpers import create_test_corpus, create_test_labels

def test_eval_accuracy():
    data = load_eval_dataset()
    results = [classify(d["input"])[0] for d in data]
    assert len(results) == len(data)

def test_eval_batch():
    corpus = create_test_corpus()
    results = batch_classify(corpus)
    assert len(results) == len(corpus)
