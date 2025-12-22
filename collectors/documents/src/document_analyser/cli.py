from __future__ import annotations

import argparse
import os
import sys
from pathlib import Path

from .config import ConfigError, ModelConfig
from .evidence_store import (
    EvidenceStoreClient,
    EvidenceStoreConfig,
    EvidenceStoreError,
    build_test_evidence_payload,
)
from .llm import LLMClient
from .pipeline import (
    DocumentAnalysisPipeline,
    DocumentLoader,
    EvidenceExtractor,
    EvidencePublisher,
)
from .requirements import get_requirement, list_requirements


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Extract structured compliance evidence from documents using OpenAI."
    )
    parser.add_argument(
        "files",
        nargs="*",
        help="Path(s) to the documents to analyse.",
    )
    parser.add_argument(
        "--focus",
        help="Optional focus area or question to guide the extraction.",
    )
    parser.add_argument("--model", help="Model name (local or remote).")
    parser.add_argument("--base-url", dest="base_url", help="OpenAI-compatible base URL.")
    parser.add_argument(
        "--api-key",
        dest="api_key",
        help="API key. For local no-auth servers set any placeholder.",
    )
    parser.add_argument(
        "--max-items",
        type=int,
        default=8,
        help="Maximum number of evidence items to return.",
    )
    parser.add_argument(
        "--push-evidence",
        dest="push_evidence",
        action="store_true",
        help="Send generated evidence to the evidence store.",
    )
    parser.add_argument(
        "--evidence-url",
        dest="evidence_url",
        help="Evidence store base URL (defaults to EVIDENCE_STORE_BASE or CONFIRMATE_API_BASE).",
    )
    parser.add_argument(
        "--evidence-auth",
        dest="evidence_auth",
        help="Evidence store credentials in client_id:client_secret format.",
    )
    parser.add_argument(
        "--evidence-target-id",
        dest="evidence_target_id",
        help="Target of evaluation ID override.",
    )
    parser.add_argument(
        "--evidence-tool-id",
        dest="evidence_tool_id",
        help="Tool identifier override.",
    )
    parser.add_argument(
        "--test-evidence",
        action="store_true",
        help="Send a minimal test evidence to verify evidence store connectivity, then exit.",
    )
    parser.add_argument(
        "--test-requirement",
        help="Run a single requirement by ID from the predefined list instead of full analysis.",
    )
    parser.add_argument(
        "--all-requirements",
        action="store_true",
        help="Run all predefined requirements instead of the general extractor.",
    )
    return parser.parse_args(argv)


def build_config(args: argparse.Namespace) -> ModelConfig:
    config = ModelConfig.from_env()
    if args.model:
        config.model = args.model
    if args.base_url:
        config.base_url = args.base_url
    if args.api_key:
        config.api_key = args.api_key
    return config


def build_evidence_config(args: argparse.Namespace) -> EvidenceStoreConfig:
    config = EvidenceStoreConfig.from_env()
    if args.evidence_url:
        config.base_url = args.evidence_url
    if args.evidence_auth:
        if ":" not in args.evidence_auth:
            raise ConfigError("Evidence auth must be in client_id:client_secret format.")
        client_id, client_secret = args.evidence_auth.split(":", 1)
        config.client_id = client_id
        config.client_secret = client_secret
    if args.evidence_target_id:
        config.target_of_evaluation_id = args.evidence_target_id
    if args.evidence_tool_id:
        config.tool_id = args.evidence_tool_id
    return config


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    if not args.test_evidence and not args.files:
        sys.stderr.write("No files provided. Specify one or more files, or use --test-evidence.\n")
        return 2
    try:
        config = build_config(args)
    except ConfigError as exc:
        sys.stderr.write(f"Configuration error: {exc}\n")
        return 1

    # Optional connectivity test for evidence store
    if args.test_evidence:
        try:
            evidence_config = build_evidence_config(args)
            payload = build_test_evidence_payload(evidence_config)
            with EvidenceStoreClient(evidence_config) as client:
                print(f"payload: {payload}")
                client.send_evidence(payload)
            sys.stderr.write("Test evidence sent successfully.\n")
            return 0
        except EvidenceStoreError as exc:
            sys.stderr.write(f"Evidence store error during test: {exc}\n")
            return 1
        except Exception as exc:  # pragma: no cover - defensive
            sys.stderr.write(f"Unexpected error during test evidence send: {exc}\n")
            return 1

    doc_paths: list[Path] = []
    for path in args.files:
        target = Path(path)
        if not target.exists():
            sys.stderr.write(f"Document not found: {target}\n")
            return 1
        doc_paths.append(target)

    loader = DocumentLoader()
    extractor = EvidenceExtractor(llm=LLMClient(config))
    auto_push = os.getenv("EVIDENCE_AUTO_PUSH", "").lower() in {"1", "true", "yes", "on"}
    push_enabled = args.push_evidence or auto_push
    publisher = None
    if push_enabled:
        publisher = EvidencePublisher(build_evidence_config(args))
    pipeline = DocumentAnalysisPipeline(loader, extractor, publisher)
    if args.test_requirement:
        requirement = get_requirement(args.test_requirement)
        if not requirement:
            sys.stderr.write(f"Unknown requirement ID: {args.test_requirement}\n")
            return 1
        requirements = [requirement]
    else:
        # Default to running all predefined requirements.
        requirements = list_requirements()

    try:
        result, pushed = pipeline.run(
            doc_paths,
            focus=args.focus,
            max_items=args.max_items,
            requirements=requirements,
            push=push_enabled,
        )
        print(result.to_json())
    except EvidenceStoreError as exc:
        sys.stderr.write(f"Evidence store error: {exc}\n")
        return 1
    except FileNotFoundError as exc:
        sys.stderr.write(f"{exc}\n")
        return 1
    except Exception as exc:  # pragma: no cover - defensive
        sys.stderr.write(f"Unexpected error during analysis: {exc}\n")
        return 1

    if push_enabled:
        if pushed:
            sys.stderr.write(f"Pushed {pushed} evidence item(s) to the evidence store.\n")
        else:
            sys.stderr.write("No evidence items produced; skipping evidence store push.\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
