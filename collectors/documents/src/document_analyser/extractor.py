from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Dict, List, Sequence

from .evidence_profiles import get_default_resource_types, normalize_resource_type
from .llm import LLMClient
from .loaders import Document, concatenate_documents
from .prompts import build_messages, build_requirement_messages, build_resource_messages
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

    @staticmethod
    def _parse_json_object(raw_response: str) -> Dict[str, Any]:
        parsed: Dict[str, Any]
        try:
            parsed = json.loads(raw_response)
        except json.JSONDecodeError:
            parsed = {
                "document_summary": "",
                "evidence": [],
                "gaps": [],
                "parse_error": "Response was not valid JSON.",
            }
        return parsed

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
        parsed = self._parse_json_object(raw_response)

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

    def analyse_resources(self, documents: Sequence[Document]) -> List[Dict[str, Any]]:
        return self.analyse_resources_with_scope(documents, include_all_resource_types=False)

    def analyse_resources_with_scope(
        self,
        documents: Sequence[Document],
        include_all_resource_types: bool = False,
    ) -> List[Dict[str, Any]]:
        """Extract ontology-backed resource evidence from documents."""
        if not documents:
            raise ValueError("No documents provided for analysis.")

        allowed_resource_types = None
        if not include_all_resource_types:
            allowed_resource_types = set(get_default_resource_types())

        normalized_items: List[Dict[str, Any]] = []
        for doc in documents:
            messages = build_resource_messages(
                document_text=doc.content,
                source_name=doc.name,
                include_all_resource_types=include_all_resource_types,
            )
            raw_response = self.llm.chat(messages, response_format="json")
            parsed = self._parse_json_object(raw_response)
            raw_items = parsed.get("resourceEvidence") or parsed.get("resources") or []
            if not isinstance(raw_items, list):
                continue

            for item in raw_items:
                if not isinstance(item, dict):
                    continue

                resource_type = normalize_resource_type(item.get("resourceType"))
                resource_wrapper = item.get("resource")
                if not resource_type and isinstance(resource_wrapper, dict) and len(resource_wrapper) == 1:
                    resource_type = normalize_resource_type(next(iter(resource_wrapper)))
                if not resource_type:
                    continue
                if allowed_resource_types is not None and resource_type not in allowed_resource_types:
                    continue

                resource_body: Dict[str, Any]
                if (
                    isinstance(resource_wrapper, dict)
                    and resource_type in resource_wrapper
                    and isinstance(resource_wrapper[resource_type], dict)
                ):
                    resource_body = dict(resource_wrapper[resource_type])
                elif isinstance(item.get("resourceBody"), dict):
                    resource_body = dict(item["resourceBody"])
                elif isinstance(resource_wrapper, dict):
                    resource_body = dict(resource_wrapper)
                else:
                    continue

                normalized_items.append(
                    {
                        "resourceType": resource_type,
                        "resource": {resource_type: resource_body},
                        "snippet": item.get("snippet") or "",
                        "citation": item.get("citation") or "",
                        "sourcePath": str(doc.path),
                    }
                )

        return normalized_items
