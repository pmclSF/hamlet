from pydantic import BaseModel


class UserProfile(BaseModel):
    full_name: str
    email: str
    account_id: str
