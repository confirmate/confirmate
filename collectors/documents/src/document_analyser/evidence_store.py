from __future__ import annotations

import os
from dataclasses import dataclass
from datetime import datetime, timezone
from typing import Any, Dict, List, Sequence
from uuid import uuid4

import httpx

from .extractor import AnalysisResult
from .loaders import Document


class EvidenceStoreError(RuntimeError):
    """Raised when evidence store interactions fail."""


@dataclass
class EvidenceStoreConfig:
    base_url: str
    token_endpoint: str
    client_id: str
    client_secret: str
    tool_id: str
    target_of_evaluation_id: str
    evidence_path: str = "/v1/evidence_store/evidence"

    @classmethod
    def from_env(cls) -> "EvidenceStoreConfig":
        base_url = os.getenv("EVIDENCE_STORE_BASE") or os.getenv(
            "CONFIRMATE_API_BASE", "http://localhost:8080"
        )
        token_endpoint = os.getenv(
            "AUTH_TOKEN_ENDPOINT", f"{base_url.rstrip('/')}/v1/auth/token"
        )
        client_id = os.getenv("AUTH_CLIENT_ID", "clouditor")
        client_secret = os.getenv("AUTH_CLIENT_SECRET", "clouditor")
        tool_id = os.getenv("EVIDENCE_TOOL_ID", "document-analyser")
        target_of_evaluation_id = os.getenv(
            "TARGET_OF_EVALUATION_ID", "00000000-0000-0000-0000-000000000000"
        )
        evidence_path = os.getenv("EVIDENCE_PATH", "/v1/evidence_store/evidence")

        return cls(
            base_url=base_url,
            token_endpoint=token_endpoint,
            client_id=client_id,
            client_secret=client_secret,
            tool_id=tool_id,
            target_of_evaluation_id=target_of_evaluation_id,
            evidence_path=evidence_path,
        )

    @property
    def evidence_url(self) -> str:
        return f"{self.base_url.rstrip('/')}{self.evidence_path}"


def _strip_none(obj: Any) -> Any:
    """Recursively remove null/empty values to keep payloads tidy."""
    if isinstance(obj, dict):
        return {k: _strip_none(v) for k, v in obj.items() if v not in (None, {}, [])}
    if isinstance(obj, list):
        return [v for v in (_strip_none(v) for v in obj) if v not in (None, {}, [])]
    return obj


def build_evidence_payloads(
    result: AnalysisResult,
    documents: Sequence[Document],
    config: EvidenceStoreConfig,
) -> List[Dict[str, Any]]:
    """Convert analysis output into evidence store payloads."""
    evidence_items = result.data.get("evidence") or []
    if not isinstance(evidence_items, list):
        return []

    payloads: List[Dict[str, Any]] = []

    for idx, item in enumerate(evidence_items):
        doc = documents[idx % len(documents)]
        response_field = item.get("responseField") or "requirementMet"
        fulfilled = item.get("fulfilled")
        snippet = (
            item.get("snippet")
            or item.get("citation")
            or item.get("evidence")
            or item.get("statement")
            or ""
        )
        citation = item.get("citation") or ""
        raw = snippet
        if citation:
            raw = f"{snippet}\nCitation: {citation}" if snippet else f"Citation: {citation}"
        evidence_body: Dict[str, Any] = {
            "id": str(uuid4()),
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "targetOfEvaluationId": config.target_of_evaluation_id,
            "toolId": config.tool_id,
            "resource": {
                "genericDocument": {
                    "id": f"{config.tool_id}:document:{idx}",
                    "name": doc.name,
                    "description": item.get("title") or "Document evidence",
                    "filetype": doc.path.suffix.lstrip("."),
                    "dataLocation": {
                        "localDataLocation": {"path": str(doc.path)}
                    },
                    response_field: fulfilled,
                    # Store the proving snippet in raw to align with proto genericDocument.
                    "raw": raw,
                }
            },
            "experimentalRelatedResourceIds": [],
        }
        payloads.append(_strip_none(evidence_body))

    return payloads


def build_test_evidence_payload(config: EvidenceStoreConfig) -> Dict[str, Any]:
    """Construct a minimal evidence payload to validate connectivity."""
    evidence_body: Dict[str, Any] = {
        "id": str(uuid4()),
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "targetOfEvaluationId": config.target_of_evaluation_id,
        "toolId": config.tool_id,
        "resource": {
            "genericDocument": {
                "id": f"{config.tool_id}:test-resource",
                "name": "evidence-store-ping",
                "description": "Connectivity test payload",
                "dataLocation": {"localDataLocation": {"path": "N/A"}},
                "raw": "ping",
            }
        },
        "experimentalRelatedResourceIds": [],
    }
    return _strip_none(evidence_body)


class EvidenceStoreClient:
    """Handles OAuth and submission to the evidence store."""

    def __init__(self, config: EvidenceStoreConfig, timeout: float = 15.0):
        self.config = config
        self._client = httpx.Client(timeout=timeout)
        self._token: str | None = None

    def _get_token(self) -> str:
        if self._token:
            return self._token

        try:
            response = self._client.post(
                self.config.token_endpoint,
                data={"grant_type": "client_credentials"},
                auth=(self.config.client_id, self.config.client_secret),
            )
        except httpx.HTTPError as exc:
            raise EvidenceStoreError(f"OAuth token request failed: {exc}") from exc

        if response.status_code >= 400:
            raise EvidenceStoreError(
                f"OAuth token request failed: {response.status_code} {response.text}"
            )

        payload = response.json()
        token = payload.get("access_token")
        if not token:
            raise EvidenceStoreError("OAuth token response missing access_token.")

        self._token = token
        return token

    def send_evidence(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        token = self._get_token()
        try:
            response = self._client.post(
                self.config.evidence_url,
                json=payload,
                headers={"Authorization": f"Bearer {token}"},
            )
        except httpx.HTTPError as exc:
            raise EvidenceStoreError(
                f"Failed to send evidence to {self.config.evidence_url}: {exc}"
            ) from exc

        if response.status_code >= 400:
            raise EvidenceStoreError(
                f"Evidence store rejected payload ({response.status_code}) "
                f"at {self.config.evidence_url}: {response.text}"
            )
        if response.status_code == 204:
            return {}
        return response.json()

    def send_batch(self, payloads: Sequence[Dict[str, Any]]) -> List[Dict[str, Any]]:
        results: List[Dict[str, Any]] = []
        for payload in payloads:
            results.append(self.send_evidence(payload))
        return results

    def close(self) -> None:
        self._client.close()

    def __enter__(self) -> "EvidenceStoreClient":
        return self

    def __exit__(self, exc_type, exc, tb) -> None:
        self.close()
