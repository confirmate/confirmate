from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, List, Sequence

from .evidence_store import (
    EvidenceStoreClient,
    EvidenceStoreConfig,
    build_evidence_payloads,
)
from .extractor import AnalysisResult, DocumentAnalyser
from .llm import LLMClient
from .loaders import Document, load_any_document
from .requirements import RequirementPrompt


@dataclass
class DocumentLoader:
    """Load documents from disk using format-aware loaders."""

    encoding: str = "utf-8"

    def load(self, paths: Iterable[str | Path]) -> List[Document]:
        documents: List[Document] = []
        for path in paths:
            documents.append(load_any_document(path, encoding=self.encoding))
        return documents


class EvidenceExtractor:
    """Build prompts and parse responses into AnalysisResult objects."""

    def __init__(self, llm: LLMClient):
        self.analyser = DocumentAnalyser(llm)

    def extract(
        self,
        documents: Sequence[Document],
        focus: str | None = None,
        max_items: int = 8,
        requirements: Sequence[RequirementPrompt] | None = None,
    ) -> AnalysisResult:
        if requirements:
            evidence_items = self.analyser.analyse_requirements(documents, requirements)
            data = {"document_summary": "", "evidence": evidence_items, "gaps": []}
            sources = [str(doc.path) for doc in documents]
            return AnalysisResult(data=data, raw_response="", sources=sources)

        return self.analyser.analyse(documents, focus=focus, max_items=max_items)


class EvidencePublisher:
    """Send evidence payloads to the evidence store."""

    def __init__(self, config: EvidenceStoreConfig):
        self.config = config

    def push(self, result: AnalysisResult, documents: Sequence[Document]) -> int:
        payloads = build_evidence_payloads(result, documents, self.config)
        if not payloads:
            return 0

        with EvidenceStoreClient(self.config) as client:
            client.send_batch(payloads)

        return len(payloads)


class DocumentAnalysisPipeline:
    """End-to-end pipeline: load -> extract -> optional push."""

    def __init__(
        self,
        loader: DocumentLoader,
        extractor: EvidenceExtractor,
        publisher: EvidencePublisher | None = None,
    ):
        self.loader = loader
        self.extractor = extractor
        self.publisher = publisher

    def run(
        self,
        paths: Sequence[Path],
        focus: str | None = None,
        max_items: int = 8,
        requirements: Sequence[RequirementPrompt] | None = None,
        push: bool = False,
    ) -> tuple[AnalysisResult, int]:
        documents = self.loader.load(paths)
        result = self.extractor.extract(
            documents,
            focus=focus,
            max_items=max_items,
            requirements=requirements,
        )

        pushed = 0
        if push and self.publisher:
            pushed = self.publisher.push(result, documents)

        return result, pushed
