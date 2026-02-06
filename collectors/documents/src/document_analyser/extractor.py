from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Dict, List, Sequence

from .llm import LLMClient
from .loaders import Document, concatenate_documents
from .prompts import build_messages, build_requirement_messages
from .requirements import RequirementPrompt


@dataclass
class AnalysisResult:
    data: Dict[str, Any]
    raw_response: str
    sources: List[str]

    def to_json(self, indent: int = 2) -> str:
        return json.dumps(
            {
                "sources": self.sources,
                "analysis": self.data,
                "raw_response": self.raw_response,
            },
            indent=indent,
        )


class DocumentAnalyser:
    """Run the LLM over one or more documents and parse the output."""

    def __init__(self, llm: LLMClient):
        self.llm = llm

    def analyse(
        self,
        documents: Sequence[Document],
        focus: str | None = None,
        max_items: int = 8,
    ) -> AnalysisResult:
        if not documents:
            raise ValueError("No documents provided for analysis.")

        merged_text = concatenate_documents(documents)
        source_name = ", ".join(doc.name for doc in documents)
        messages = build_messages(
            document_text=merged_text,
            source_name=source_name,
            focus=focus,
            max_items=max_items,
        )

        raw_response = self.llm.chat(messages, response_format="json")
        parsed: Dict[str, Any]
        try:
            parsed = json.loads(raw_response)
        except json.JSONDecodeError:
            # Keep the raw response for debugging instead of failing hard.
            parsed = {
                "document_summary": "",
                "evidence": [],
                "gaps": [],
                "parse_error": "Response was not valid JSON.",
            }

        sources = [str(doc.path) for doc in documents]
        return AnalysisResult(data=parsed, raw_response=raw_response, sources=sources)

    def analyse_requirements(
        self,
        documents: Sequence[Document],
        requirements: Sequence[RequirementPrompt],
    ) -> List[Dict[str, Any]]:
        """Run one evidence extraction per requirement."""
        if not documents:
            raise ValueError("No documents provided for analysis.")
        merged_text = concatenate_documents(documents)
        source_name = ", ".join(doc.name for doc in documents)
        evidence_items: List[Dict[str, Any]] = []

        def _parse_bool(value: Any) -> bool | None:
            if isinstance(value, bool):
                return value
            if isinstance(value, str):
                normalized = value.strip().lower()
                if normalized in {"true", "yes", "y", "1", "fulfilled"}:
                    return True
                if normalized in {"false", "no", "n", "0", "unfulfilled", "not fulfilled"}:
                    return False
            return None

        for requirement in requirements:
            messages = build_requirement_messages(
                document_text=merged_text,
                requirement=requirement,
                source_name=source_name,
            )
            raw_response = self.llm.chat(messages, response_format="json")
            try:
                parsed = json.loads(raw_response)
            except json.JSONDecodeError:
                parsed = {}
            field_name = getattr(requirement, "response_field_name", None) or "requirementMet"
            fulfilled = _parse_bool(parsed.get(field_name)) or _parse_bool(parsed.get("fulfilled"))
            if fulfilled is None:
                fulfilled = False
            evidence_items.append(
                {
                    "title": requirement.name,
                    "evidence": parsed.get("statement") or parsed.get("evidence") or "",
                    "snippet": parsed.get("snippet") or "",
                    "citation": parsed.get("citation") or "",
                    "fulfilled": fulfilled,
                    "responseField": field_name,
                    "confidence": parsed.get("confidence") or "low",
                    "requirementId": requirement.id,
                    "resourceType": requirement.resource_type,
                }
            )

        return evidence_items
