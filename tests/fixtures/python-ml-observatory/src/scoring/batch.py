"""Batch scoring pipeline."""
from src.models.classifier import classify

def score_batch(texts):
    """Score a batch of texts and return predictions."""
    return [classify(t) for t in texts]

def evaluate_batch(texts, labels):
    """Evaluate a batch against labels."""
    preds = [classify(t)[0] for t in texts]
    correct = sum(1 for p, l in zip(preds, labels) if p == l)
    return {"accuracy": correct / len(texts) if texts else 0.0, "total": len(texts)}
