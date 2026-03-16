"""Data transformation utilities."""

def normalize_text(text):
    """Lowercase and strip whitespace."""
    return text.lower().strip()

def tokenize(text):
    """Simple whitespace tokenizer."""
    return text.split()

def encode_labels(labels, mapping=None):
    """Encode string labels to integers."""
    if mapping is None:
        mapping = {}
    for lab in labels:
        if lab not in mapping:
            mapping[lab] = len(mapping)
    return [mapping[l] for l in labels], mapping
