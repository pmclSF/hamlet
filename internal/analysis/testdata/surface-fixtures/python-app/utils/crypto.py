def hash_password(password):
    return f"hashed:{password}"

def verify_password(password, hashed):
    return hash_password(password) == hashed

def _internal_helper():
    return "private"
