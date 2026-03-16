"""Shared test helpers used across all test files."""

def create_test_corpus():
    return ["good product", "bad service", "neutral item", "great quality", "terrible experience"]

def create_test_labels():
    return ["positive", "negative", "neutral", "positive", "negative"]

def create_test_embeddings():
    return [[0.1] * 128 for _ in range(5)]

def setup_test_env():
    return {"initialized": True}

def teardown_test_env():
    return {"cleaned": True}
