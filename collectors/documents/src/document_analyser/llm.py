from __future__ import annotations

from typing import Dict, List, Optional

from openai import OpenAI

from .config import ModelConfig


class LLMClient:
    """Thin wrapper around the OpenAI chat completion API."""

    def __init__(self, config: ModelConfig):
        self.config = config
        self._client = OpenAI(api_key=config.api_key, base_url=config.base_url)

    def chat(
        self,
        messages: List[Dict[str, str]],
        response_format: Optional[str] = None,
    ) -> str:
        params: Dict[str, object] = {
            "model": self.config.model,
            "messages": messages,
            "temperature": self.config.temperature,
            "max_tokens": self.config.max_tokens,
        }
        if response_format == "json":
            params["response_format"] = {"type": "json_object"}

        response = self._client.chat.completions.create(**params)
        message = response.choices[0].message
        return message.content or ""
