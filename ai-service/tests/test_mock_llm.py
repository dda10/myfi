"""Tests for the mock LLM and test mode infrastructure.

Requirements: 46.3
"""

from ezistock_ai.llm.mock import MockChatModel, _select_response
from ezistock_ai.llm.router import LLMRouter, TaskType


def test_mock_chat_model_returns_string():
    """MockChatModel should return a non-empty string response."""
    from langchain_core.messages import HumanMessage

    model = MockChatModel()
    result = model.invoke([HumanMessage(content="Hello")])
    assert result.content
    assert isinstance(result.content, str)


def test_mock_selects_technical_response():
    """Keywords like 'technical' should trigger the technical canned response."""
    resp = _select_response("Analyze the technical indicators for FPT")
    assert "composite_signal" in resp
    assert "rsi" in resp


def test_mock_selects_news_response():
    """Keywords like 'news' should trigger the news canned response."""
    resp = _select_response("What is the latest news sentiment?")
    assert "sentiment" in resp
    assert "catalysts" in resp


def test_mock_selects_recommendation_response():
    """Keywords like 'recommend' should trigger the recommendation response."""
    resp = _select_response("Give me an investment recommendation")
    assert "action" in resp
    assert "target_price" in resp


def test_mock_selects_strategy_response():
    """Keywords like 'strategy' should trigger the strategy response."""
    resp = _select_response("Build a trading strategy with entry and stop_loss")
    assert "signal_direction" in resp
    assert "entry_price" in resp


def test_mock_default_response():
    """Unknown input should return the default mock response."""
    resp = _select_response("random unrelated text")
    assert "mock AI response" in resp


def test_llm_router_test_mode(test_config):
    """LLMRouter in test mode should return MockChatModel instances."""
    router = LLMRouter(test_config)
    model = router.get_model(TaskType.EXTRACTION)
    assert model._llm_type == "mock"


def test_llm_router_all_task_types(test_config):
    """All task types should work in test mode."""
    router = LLMRouter(test_config)
    for task_type in TaskType:
        model = router.get_model(task_type)
        assert model._llm_type == "mock"
