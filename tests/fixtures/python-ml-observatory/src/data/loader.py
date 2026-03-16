"""Data loading utilities."""

def load_training_data(path):
    """Load training dataset from file path."""
    return [{"text": "sample", "label": "pos"}]

def load_eval_dataset():
    """Load evaluation dataset."""
    return [
        {"input": "good product", "expected": "positive"},
        {"input": "terrible", "expected": "negative"},
        {"input": "it works", "expected": "neutral"},
    ]

def split_dataset(data, ratio=0.8):
    """Split data into train and test sets."""
    idx = int(len(data) * ratio)
    return data[:idx], data[idx:]
