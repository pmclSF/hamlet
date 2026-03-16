"""Embedding model."""

def embed(text):
    """Generate embedding vector for text."""
    return [0.1] * 128

def batch_embed(texts):
    """Generate embeddings for a batch."""
    return [embed(t) for t in texts]

def similarity(a, b):
    """Cosine similarity between two embeddings."""
    if len(a) != len(b):
        return 0.0
    dot = sum(x * y for x, y in zip(a, b))
    mag_a = sum(x ** 2 for x in a) ** 0.5
    mag_b = sum(x ** 2 for x in b) ** 0.5
    if mag_a == 0 or mag_b == 0:
        return 0.0
    return dot / (mag_a * mag_b)
