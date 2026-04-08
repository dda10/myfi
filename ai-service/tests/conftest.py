"""Shared test fixtures for the EziStock AI Service.

Sets EZISTOCK_TEST_MODE=true so all LLM calls use MockChatModel.
Requirements: 46.3
"""

import os

import pytest

# Force test mode before any config is loaded
os.environ["EZISTOCK_TEST_MODE"] = "true"
os.environ["EZISTOCK_OPENAI_API_KEY"] = "test-key"


@pytest.fixture
def test_config():
    """Return a Config instance with test mode enabled."""
    from ezistock_ai.config import Config

    return Config(test_mode=True, openai_api_key="test-key")


@pytest.fixture
def mock_llm_router(test_config):
    """Return an LLMRouter that uses MockChatModel for all task types."""
    from ezistock_ai.llm.router import LLMRouter

    return LLMRouter(test_config)
