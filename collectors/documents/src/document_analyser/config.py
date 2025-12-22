from __future__ import annotations

import os
from dataclasses import dataclass


DEFAULT_MODEL = "gpt-4o-mini"
DEFAULT_TEMPERATURE = 0.0
DEFAULT_MAX_TOKENS = 800


class ConfigError(RuntimeError):
    """Raised when configuration is missing or invalid."""


@dataclass
class ModelConfig:
    api_key: str
    model: str = DEFAULT_MODEL
    temperature: float = DEFAULT_TEMPERATURE
    max_tokens: int = DEFAULT_MAX_TOKENS
    base_url: str | None = None

    @classmethod
    def from_env(cls) -> "ModelConfig":
        """Load configuration from environment variables."""
        api_key = os.getenv("DOC_ANALYSER_API_KEY") or os.getenv("OPENAI_API_KEY")
        base_url = os.getenv("DOC_ANALYSER_BASE_URL") or os.getenv("OPENAI_BASE_URL")
        if not api_key and not base_url:
            raise ConfigError(
                "API key missing. Set OPENAI_API_KEY (or DOC_ANALYSER_API_KEY) or provide a DOC_ANALYSER_BASE_URL/OPENAI_BASE_URL for a local model."
            )
        if not api_key:
            # Some local OpenAI-compatible servers do not require a key.
            api_key = "not-set"

        model = os.getenv("DOC_ANALYSER_MODEL", DEFAULT_MODEL)
        temperature = float(os.getenv("DOC_ANALYSER_TEMPERATURE", DEFAULT_TEMPERATURE))
        max_tokens = int(os.getenv("DOC_ANALYSER_MAX_TOKENS", DEFAULT_MAX_TOKENS))

        return cls(
            api_key=api_key,
            model=model,
            temperature=temperature,
            max_tokens=max_tokens,
            base_url=base_url,
        )
