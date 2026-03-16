"""Text classifier model."""

def classify(text):
    """Classify text sentiment."""
    text_lower = text.lower()
    if any(w in text_lower for w in ["good", "great", "excellent"]):
        return "positive", 0.9
    if any(w in text_lower for w in ["bad", "terrible", "awful"]):
        return "negative", 0.85
    return "neutral", 0.6

def batch_classify(texts):
    """Classify a batch of texts."""
    return [classify(t) for t in texts]
