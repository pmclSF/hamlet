"""Evaluation metrics for model assessment."""


def accuracy(predictions, labels):
    """Compute classification accuracy."""
    if not predictions:
        return 0.0
    correct = sum(1 for p, l in zip(predictions, labels) if p == l)
    return correct / len(predictions)


def precision_recall(predictions, labels, target_class):
    """Compute precision and recall for a target class."""
    tp = sum(1 for p, l in zip(predictions, labels) if p == target_class and l == target_class)
    fp = sum(1 for p, l in zip(predictions, labels) if p == target_class and l != target_class)
    fn = sum(1 for p, l in zip(predictions, labels) if p != target_class and l == target_class)

    precision = tp / (tp + fp) if (tp + fp) > 0 else 0.0
    recall = tp / (tp + fn) if (tp + fn) > 0 else 0.0
    return precision, recall


def f1_score(precision, recall):
    """Compute F1 score from precision and recall."""
    if precision + recall == 0:
        return 0.0
    return 2 * (precision * recall) / (precision + recall)
