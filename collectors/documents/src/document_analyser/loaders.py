from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Iterable, List

from pypdf import PdfReader


@dataclass
class Document:
    path: Path
    content: str

    @property
    def name(self) -> str:
        return self.path.name


def load_text_document(path: str | Path, encoding: str = "utf-8") -> Document:
    """Read a text-based document from disk with a forgiving decoder."""
    target = Path(path)
    if not target.exists():
        raise FileNotFoundError(f"Document not found: {target}")

    try:
        text = target.read_text(encoding=encoding)
    except UnicodeDecodeError:
        # Fallback for mixed encodings; binary content will be filtered by caller.
        text = target.read_text(encoding=encoding, errors="ignore")

    return Document(path=target, content=text)


def load_pdf_document(path: str | Path) -> Document:
    """Extract text from a PDF using pypdf."""
    target = Path(path)
    if not target.exists():
        raise FileNotFoundError(f"Document not found: {target}")

    reader = PdfReader(str(target))
    pages: List[str] = []
    for idx, page in enumerate(reader.pages, start=1):
        try:
            text = page.extract_text() or ""
        except Exception:
            text = ""
        if text.strip():
            pages.append(f"[Page {idx}]\n{text.strip()}")

    content = "\n\n".join(pages).strip() if pages else "[No extractable text found in PDF]"
    return Document(path=target, content=content)


def load_document(path: str | Path, encoding: str = "utf-8") -> Document:
    """Backward-compatible alias for text loader."""
    return load_text_document(path, encoding=encoding)


def load_any_document(path: str | Path, encoding: str = "utf-8") -> Document:
    """Dispatch to the appropriate loader based on file extension."""
    target = Path(path)
    suffix = target.suffix.lower()
    if suffix == ".pdf":
        return load_pdf_document(target)
    return load_text_document(target, encoding=encoding)


def concatenate_documents(documents: Iterable[Document]) -> str:
    """Join multiple documents into one prompt-friendly string."""
    parts = []
    for doc in documents:
        header = f"### Document: {doc.name}\n"
        parts.append(header + doc.content.strip() + "\n")
    return "\n".join(parts).strip()
