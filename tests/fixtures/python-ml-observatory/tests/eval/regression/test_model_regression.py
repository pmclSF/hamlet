"""Model regression evaluation."""
from src.models.classifier import classify
from src.models.embeddings import embed, similarity
from src.data.loader import load_eval_dataset

def test_classifier_consistency():
    label1, conf1 = classify("great product")
    label2, conf2 = classify("great product")
    assert label1 == label2
    assert conf1 == conf2

def test_embedding_consistency():
    v1 = embed("hello")
    v2 = embed("hello")
    assert similarity(v1, v2) > 0.99

def test_eval_dataset_stable():
    data = load_eval_dataset()
    assert len(data) == 3
