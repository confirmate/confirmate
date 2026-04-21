from __future__ import annotations

from typing import List, Dict

from .evidence_profiles import render_resource_profiles_for_prompt


def build_messages(
    document_text: str,
    source_name: str | None = None,
    focus: str | None = None,
    max_items: int = 8,
) -> List[Dict[str, str]]:
    """Compose the chat messages for the LLM."""
    context_name = source_name or "document"
    user_instructions = f"Source: {context_name}.\n\n"
    if focus:
        user_instructions += f"Focus areas provided by user: {focus}\n\n"

    user_instructions += (
        "Extract evidence and return the JSON object described in the system message.\n"
        "Only include claims supported by the text."
    )

    system_prompt = """
    You are Document-Analyser, a careful assistant that extracts compliance evidence from unstructured documents
    such as source code, security policies, and reports. Work step by step, focus on verifiable statements, and
    prefer quoting the original phrasing where possible.

    Produce a compact JSON object with these keys:
    - document_summary: 2 sentence overview of the document content and purpose.
    - evidence: list of up to {max_items} evidence objects. Each evidence object must contain:
    - title: short label for the requirement, control, or claim.
    - evidence: concise statement derived from the document.
    - snippet: verbatim quote or text excerpt used to support the evidence (keep it short).
    - citation: page number hint for the snippet, e.g., "Page 3". Leave empty if unknown.
    - confidence: one of ["high", "medium", "low"] describing certainty.
    - gaps: list of missing information or unresolved questions, may be empty.

    Keep the JSON machine-readable and avoid markdown.
    """
    system_prompt = system_prompt.format(max_items=max_items)

    return [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": user_instructions + "\n\n" + document_text},
    ]


def build_requirement_messages(
    document_text: str,
    requirement,
    source_name: str | None = None,
) -> List[Dict[str, str]]:
    """Compose messages for a single requirement, requesting an evidence snippet."""
    context_name = source_name or "document"
    field_name = getattr(requirement, "response_field_name", None) or "requirementMet"
    schema_hint = getattr(requirement, "response_schema", None) or {
        "requirementId": requirement.id,
        field_name: True,
        "snippet": "Verbatim quote or text excerpt proving the requirement",
        "citation": "Page number, e.g., 'Page 3' (empty if unknown)",
        "confidence": "high|medium|low",
    }
    system_prompt = (
        "You are Document-Analyser. Check if the document contains the required information.\n"
        "Return a JSON object (not an array) with these fields:\n"
        f"{schema_hint}\n"
        "Use camelCase keys. Do not include markdown. Set the boolean field to true if the document contains the required information, otherwise false. "
        "Snippet must be the quote/excerpt you used. "
        'Citation must be the page number (e.g., "Page 2") if available, otherwise empty. '
        'If no evidence exists, set the boolean field to false, snippet and citation to empty strings, and confidence to "low".'
    )
    user_prompt = (
        f"Source: {context_name}.\n"
        f"Requirement: {getattr(requirement, 'name', requirement.id)} ({requirement.id}).\n"
        f"Instruction: {requirement.prompt}\n\n"
        "Return only the JSON object."
    )
    return [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": user_prompt + "\n\n" + document_text},
    ]


def build_resource_messages(
    document_text: str,
    source_name: str | None = None,
    max_items: int = 20,
    include_all_resource_types: bool = False,
) -> List[Dict[str, str]]:
    """Compose messages for extracting ontology-backed resource evidence."""
    context_name = source_name or "document"
    supported_profiles = render_resource_profiles_for_prompt(
        include_all=include_all_resource_types
    )
    scope_hint = (
        "all ontology-backed resource types from the proto schema"
        if include_all_resource_types
        else "the default CRA-focused resource types only"
    )
    system_prompt = f"""
    You are Document-Analyser. Extract ontology-backed evidence resources from the document.
    Return a compact JSON object with exactly one top-level key named "resourceEvidence".
    The value of "resourceEvidence" must be an array with at most {max_items} objects.

    Each item in the array must have:
    - resourceType: one of the supported resource types listed below.
    - resource: an object with exactly one top-level key equal to resourceType.
      The nested object must use camelCase field names and include only fields that exist in the proto-derived schema for that resource type.
    - snippet: a short verbatim quote or excerpt from the document proving the resource fields.
    - citation: a page hint such as "Page 2". Use the [Page N] markers from the document text when available.

    Supported resource types:
    {supported_profiles}

    Rules:
    - Restrict extraction to {scope_hint}.
    - Only extract resources explicitly supported by the document.
    - Populate as many proto-defined fields as the document supports, not just id and name.
    - Do not invent relationships, timestamps, contact details, labels, or metadata that are not grounded in the document.
    - Prefer stable identifiers from the document. If the document does not provide an explicit ID, derive a short slug from the resource name.
    - For repeated proto fields, return JSON arrays.
    - For message-typed proto fields, return nested JSON objects and fill their subfields when the document clearly supports them.
    - Omit fields you cannot support from the text instead of guessing.
    - Keep snippets short and evidence-based.
    - If no supported resources are present, return {{"resourceEvidence": []}}.
    - Return JSON only. Do not include markdown.
    """
    user_prompt = (
        f"Source: {context_name}.\n"
        "Extract the supported ontology resources described in this document."
    )
    return [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": user_prompt + "\n\n" + document_text},
    ]
