"""Tests for retrieval search."""
from src.retrieval.search import retrieve, rerank
from src.conftest_helpers import create_test_corpus

def test_retrieve():
    corpus = create_test_corpus()
    results = retrieve("good product", corpus, top_k=3)
    assert len(results) <= 3

def test_rerank():
    candidates = [("doc1", 0.9), ("doc2", 0.8)]
    result = rerank("query", candidates)
    assert len(result) == 2
