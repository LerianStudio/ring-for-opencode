"""Example module for testing AST extraction."""

import os
from typing import Optional, List
from dataclasses import dataclass


@dataclass
class User:
    id: int
    name: str


class UserService:
    """Service for managing users."""

    def __init__(self, db_url: str):
        self.db_url = db_url

    def get_user(self, user_id: int) -> Optional[User]:
        """Get a user by ID."""
        return None

    def list_users(self) -> List[User]:
        """List all users."""
        return []


def greet(name: str) -> str:
    """Return a greeting message."""
    return f"Hello, {name}!"


async def fetch_data(url: str) -> dict:
    """Fetch data from a URL."""
    return {}


def format_name(name: str) -> str:
    """Format a name."""
    return name.strip().title()
