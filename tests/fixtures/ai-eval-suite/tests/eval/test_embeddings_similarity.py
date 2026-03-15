"""Test embedding quality via similarity properties."""
from src.models.embeddings import embed, cosine_similarity


class TestEmbeddingSimilarity:
    def test_identical_texts_have_similarity_one(self):
        a = embed("hello world")
        b = embed("hello world")
        sim = cosine_similarity(a, b)
        assert abs(sim - 1.0) < 0.001

    def test_different_texts_have_lower_similarity(self):
        a = embed("the cat sat on the mat")
        b = embed("quantum physics equations")
        sim = cosine_similarity(a, b)
        assert sim < 0.9, f"Expected dissimilar texts to have low similarity, got {sim}"

    def test_embedding_dimensions(self):
        vec = embed("test text", dimensions=64)
        assert len(vec) == 64

    def test_embedding_deterministic(self):
        a = embed("reproducible")
        b = embed("reproducible")
        assert a == b

    def test_similar_texts_closer_than_random(self):
        a = embed("good product excellent quality")
        b = embed("great product amazing quality")
        c = embed("quantum physics dark matter")
        sim_ab = cosine_similarity(a, b)
        sim_ac = cosine_similarity(a, c)
        assert sim_ab > sim_ac, "Similar texts should be closer than dissimilar"
