"""Prompt building utilities."""

system_prompt = "You are a helpful ML analysis assistant."

def build_classification_prompt(text, context=None):
    """Build a classification prompt."""
    return f"Classify: {text}"

def build_safety_prompt(text):
    """Build a safety evaluation prompt."""
    return f"Evaluate safety: {text}"

def build_retrieval_prompt(query, context):
    """Build a retrieval-augmented prompt."""
    return f"Query: {query}\nContext: {context}"
