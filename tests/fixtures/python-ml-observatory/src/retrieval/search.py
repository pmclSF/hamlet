"""Retrieval and search engine."""
from src.models.embeddings import embed, similarity

def retrieve(query, corpus, top_k=5):
    """Retrieve top_k most similar documents."""
    q_emb = embed(query)
    scored = [(doc, similarity(q_emb, embed(doc))) for doc in corpus]
    scored.sort(key=lambda x: x[1], reverse=True)
    return scored[:top_k]

def rerank(query, candidates):
    """Re-rank candidates by relevance."""
    return candidates
