"""Evaluation metrics."""

def accuracy(predictions, labels):
    """Compute accuracy."""
    if not predictions:
        return 0.0
    correct = sum(1 for p, l in zip(predictions, labels) if p == l)
    return correct / len(predictions)

def precision_recall(predictions, labels, positive_label):
    """Compute precision and recall."""
    tp = sum(1 for p, l in zip(predictions, labels) if p == positive_label and l == positive_label)
    fp = sum(1 for p, l in zip(predictions, labels) if p == positive_label and l != positive_label)
    fn = sum(1 for p, l in zip(predictions, labels) if p != positive_label and l == positive_label)
    prec = tp / (tp + fp) if (tp + fp) > 0 else 0.0
    rec = tp / (tp + fn) if (tp + fn) > 0 else 0.0
    return prec, rec
