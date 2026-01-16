"""Example module for testing AST extraction."""

import logging
from typing import Optional, List, Dict
from dataclasses import dataclass, field


@dataclass
class User:
    id: int
    name: str
    email: str  # Added field
    is_active: bool = True  # Added field with default


@dataclass
class Config:  # New dataclass
    debug: bool = False
    timeout: int = 30


class UserService:
    """Service for managing users."""

    def __init__(self, db_url: str, config: Config):  # Changed signature
        self.db_url = db_url
        self.config = config

    def get_user(self, user_id: int) -> Optional[User]:
        """Get a user by ID."""
        logging.info(f"Fetching user {user_id}")  # Changed implementation
        return None

    def list_users(self, active_only: bool = False) -> List[User]:  # Changed signature
        """List all users."""
        return []

    async def update_user(self, user_id: int, data: Dict) -> User:  # New async method
        """Update a user."""
        return User(id=user_id, name="", email="")


def greet(name: str, greeting: str = "Hello") -> str:  # Added parameter
    """Return a greeting message."""
    return f"{greeting}, {name}!"


async def fetch_data(url: str, timeout: int = 30) -> dict:  # Added parameter
    """Fetch data from a URL."""
    return {}


# format_name removed


def validate_email(email: str) -> bool:  # New function
    """Validate an email address."""
    return "@" in email
