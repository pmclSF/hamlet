"""Dataset utilities for evaluation."""


EVAL_DATASET = [
    ("This product is amazing and I love it", "positive"),
    ("Terrible experience, worst purchase ever", "negative"),
    ("The item arrived on time", "neutral"),
    ("Great quality, excellent craftsmanship", "positive"),
    ("Bad quality, very disappointing", "negative"),
    ("It works as described", "neutral"),
    ("I hate this, awful product", "negative"),
    ("Good value for the price", "positive"),
    ("Not what I expected", "neutral"),
    ("The best thing I have ever bought", "positive"),
]


def load_eval_dataset():
    """Load the evaluation dataset."""
    texts = [t for t, _ in EVAL_DATASET]
    labels = [l for _, l in EVAL_DATASET]
    return texts, labels


def split_dataset(texts, labels, ratio=0.8):
    """Split dataset into train and test sets."""
    split_idx = int(len(texts) * ratio)
    return (texts[:split_idx], labels[:split_idx]), (texts[split_idx:], labels[split_idx:])
