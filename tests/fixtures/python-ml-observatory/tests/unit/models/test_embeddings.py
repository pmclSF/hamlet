"""Tests for embeddings."""
from src.models.embeddings import embed, batch_embed, similarity

def test_embed():
    vec = embed("hello")
    assert len(vec) == 128

def test_batch_embed():
    vecs = batch_embed(["hello", "world"])
    assert len(vecs) == 2

def test_similarity():
    a = embed("hello")
    b = embed("hello")
    assert similarity(a, b) > 0.9
