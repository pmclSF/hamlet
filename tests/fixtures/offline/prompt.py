# Deliberately drifted fixture for the offline-guarantee CI job.
# The `import openai` marks this file as a prompt surface; do not remove it.
import openai
from models import UserProfile


def build(user: UserProfile) -> str:
    # Seeded drift: `acount_id` misspells the schema's `account_id`, so
    # discovery, analyze, the severity gate, and the fix dry-run all have
    # real work to do when this fixture is analyzed.
    return f"""You are a support assistant.
The user id is {user.acount_id}; the name is {user.full_name}."""
