from __future__ import annotations

import re
from pathlib import Path
from typing import Any, Dict

from .evidence_profiles import get_resource_profiles, normalize_resource_type


def guess_filetype(path: str | Path) -> str:
    suffix = Path(path).suffix.lower()
    if suffix == ".pdf":
        return "application/pdf"
    if suffix == ".json":
        return "application/json"
    if suffix == ".txt":
        return "text/plain"
    return suffix.lstrip(".")


def build_resource_raw(snippet: str, citation: str, source_path: str | Path) -> str:
    parts = []
    cleaned_snippet = snippet.strip()
    cleaned_citation = citation.strip()
    if cleaned_snippet:
        parts.append(cleaned_snippet)
    if cleaned_citation:
        parts.append(f"Citation: {cleaned_citation}")
    parts.append(f"Source document: {source_path}")
    return "\n".join(parts)


def _slugify(value: str) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")
    return slug or "resource"


def _coerce_text(value: Any) -> str:
    if isinstance(value, str):
        return value.strip()
    return str(value).strip() if value is not None else ""


def _derive_name(resource_type: str, body: Dict[str, Any], snippet: str) -> str:
    existing = _coerce_text(body.get("name"))
    if existing:
        return existing

    snippet_text = _coerce_text(snippet)
    if snippet_text:
        return snippet_text.splitlines()[0][:160].strip(" .:-")

    return resource_type


def _derive_description(body: Dict[str, Any], snippet: str) -> str:
    existing = _coerce_text(body.get("description"))
    if existing:
        return existing

    snippet_text = _coerce_text(snippet)
    if snippet_text:
        return snippet_text[:500]

    return ""


def _derive_id(resource_type: str, body: Dict[str, Any], snippet: str) -> str:
    existing = _coerce_text(body.get("id"))
    if existing:
        return existing

    name = _derive_name(resource_type, body, snippet)
    return f"{resource_type}-{_slugify(name)}"


def enrich_resource_body(
    resource_type: str,
    resource_body: Dict[str, Any],
    *,
    source_path: str | Path,
    snippet: str = "",
    citation: str = "",
) -> Dict[str, Any]:
    canonical_type = normalize_resource_type(resource_type)
    if canonical_type is None:
        raise ValueError(f"Unsupported resource type: {resource_type}")

    profile = get_resource_profiles()[canonical_type]
    allowed_fields = {field.name for field in profile.fields}
    body = {
        key: value
        for key, value in dict(resource_body).items()
        if key in allowed_fields
    }

    if "name" in allowed_fields:
        body["name"] = _derive_name(canonical_type, body, snippet)
    if "id" in allowed_fields:
        body["id"] = _derive_id(canonical_type, body, snippet)
    if "description" in allowed_fields:
        description = _derive_description(body, snippet)
        if description:
            body["description"] = description

    raw_text = build_resource_raw(snippet=snippet, citation=citation, source_path=source_path)
    if raw_text and "raw" in allowed_fields:
        existing_raw = str(body.get("raw", "")).strip()
        if not existing_raw:
            body["raw"] = raw_text
        elif raw_text not in existing_raw:
            body["raw"] = f"{existing_raw}\n\n{raw_text}"

    if profile.supports_data_location and "dataLocation" not in body:
        body["dataLocation"] = {"localDataLocation": {"path": str(source_path)}}

    if profile.supports_filetype and "filetype" not in body:
        body["filetype"] = guess_filetype(source_path)

    return body
