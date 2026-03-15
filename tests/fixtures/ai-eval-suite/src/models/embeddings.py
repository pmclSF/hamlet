"""Simple embedding generation for evaluation testing."""

import hashlib


def embed(text, dimensions=128):
    """Generate a deterministic pseudo-embedding for text."""
    h = hashlib.sha256(text.encode()).hexdigest()
    values = []
    for i in range(0, min(len(h), dimensions * 2), 2):
        values.append((int(h[i:i+2], 16) - 128) / 128.0)
    while len(values) < dimensions:
        values.append(0.0)
    return values[:dimensions]


def cosine_similarity(a, b):
    """Compute cosine similarity between two vectors."""
    dot = sum(x * y for x, y in zip(a, b))
    norm_a = sum(x * x for x in a) ** 0.5
    norm_b = sum(x * x for x in b) ** 0.5
    if norm_a == 0 or norm_b == 0:
        return 0.0
    return dot / (norm_a * norm_b)
