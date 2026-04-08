"""Configuration via pydantic-settings for environment variable management."""

from pydantic_settings import BaseSettings, SettingsConfigDict


class Config(BaseSettings):
    """EziStock AI Service configuration.

    All values are read from environment variables (prefixed EZISTOCK_)
    or from a .env file in the project root.
    """

    model_config = SettingsConfigDict(
        env_prefix="EZISTOCK_",
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    # --- Server ---
    grpc_port: int = 50051
    rest_port: int = 8000
    debug: bool = False

    # --- Database ---
    database_url: str = "postgres://ezistock:ezistock_dev@localhost:5432/ezistock"

    # --- Redis ---
    redis_url: str = "redis://localhost:6379/0"

    # --- S3 / MinIO ---
    s3_endpoint: str = "http://localhost:9000"
    s3_bucket: str = "ezistock"
    s3_access_key: str = ""
    s3_secret_key: str = ""
    s3_use_path_style: bool = True  # True for MinIO, False for AWS S3

    # --- LLM Providers ---
    openai_api_key: str = ""
    anthropic_api_key: str = ""
    google_api_key: str = ""

    # --- LLM Model Names ---
    llm_lightweight_model: str = "gpt-4o-mini"  # Data extraction, classification
    llm_capable_model: str = "gpt-4o"  # Analysis, synthesis
    llm_conversational_model: str = "gpt-4o"  # User-facing chat

    # --- LLM Budget ---
    llm_daily_budget_usd: float = 10.0

    # --- Go Backend ---
    go_backend_url: str = "http://localhost:8080"

    # --- Test Mode ---
    test_mode: bool = False  # When True, use MockChatModel instead of real LLM providers
