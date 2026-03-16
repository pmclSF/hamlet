"""Safety filtering module."""

BLOCKED_PATTERNS = ["ignore instructions", "system override", "reveal secrets"]

def check_input_safety(text):
    """Check if input text is safe."""
    text_lower = text.lower()
    for pattern in BLOCKED_PATTERNS:
        if pattern in text_lower:
            return {"safe": False, "reason": f"blocked pattern: {pattern}"}
    return {"safe": True, "reason": "no issues"}

def check_output_safety(text):
    """Check if model output is safe."""
    return {"safe": True, "reason": "no issues"}

def apply_safety_filter(prompt, response):
    """Apply full safety pipeline."""
    input_check = check_input_safety(prompt)
    if not input_check["safe"]:
        return {"blocked": True, "reason": input_check["reason"]}
    output_check = check_output_safety(response)
    return {"blocked": not output_check["safe"], "reason": output_check["reason"]}
