"""Classifier accuracy evaluation."""
from src.models.classifier import classify, batch_classify
from src.data.loader import load_eval_dataset
from src.conftest_helpers import create_test_corpus, create_test_labels

def test_accuracy_above_threshold():
    data = load_eval_dataset()
    correct = 0
    for item in data:
        label, _ = classify(item["input"])
        if label == item["expected"]:
            correct += 1
    assert correct / len(data) > 0.3

def test_batch_accuracy():
    corpus = create_test_corpus()
    labels = create_test_labels()
    results = batch_classify(corpus)
    preds = [r[0] for r in results]
    correct = sum(1 for p, l in zip(preds, labels) if p == l)
    assert correct > 0
