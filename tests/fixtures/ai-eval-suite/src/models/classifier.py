"""Simple text classifier for evaluation testing."""

CATEGORIES = ['positive', 'negative', 'neutral']


def classify(text):
    """Classify text sentiment. Returns category and confidence."""
    text_lower = text.lower()
    positive_words = ['good', 'great', 'excellent', 'amazing', 'love']
    negative_words = ['bad', 'terrible', 'awful', 'hate', 'worst']

    pos_count = sum(1 for w in positive_words if w in text_lower)
    neg_count = sum(1 for w in negative_words if w in text_lower)

    if pos_count > neg_count:
        return 'positive', min(0.5 + pos_count * 0.1, 0.99)
    elif neg_count > pos_count:
        return 'negative', min(0.5 + neg_count * 0.1, 0.99)
    return 'neutral', 0.5


def batch_classify(texts):
    """Classify a batch of texts."""
    return [classify(t) for t in texts]
