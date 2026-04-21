from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any, Sequence


DEFAULT_RESOURCE_TYPES = (
    "product",
    "application",
    "contactPerson",
    "monitoringProcedure",
    "memory",
    "virtualMachine",
    "userInformationAndIntructionDocument",
    "sbomDocument",
    "euDeclarationOfConformity",
    "distributionOfUpdatesDocument",
    "productionAndMonitoringProcessDocument",
    "cyberSecurityRiskAssessmentDocument",
    "coordinatedVulnerabilityDisclosurePolicy",
)


def parse_args(argv: Sequence[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate PDF fixtures from evidence objects in test_evidence.json."
    )
    parser.add_argument(
        "--input",
        dest="input_path",
        default=str(get_default_input_path()),
        help="Path to the evidence JSON fixture.",
    )
    parser.add_argument(
        "--output-dir",
        dest="output_dir",
        default=str(get_default_output_dir()),
        help="Directory to write the generated PDFs into.",
    )
    return parser.parse_args(argv)


def get_repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


def get_default_input_path() -> Path:
    return get_repo_root() / "test_evidence.json"


def get_default_output_dir() -> Path:
    return Path(__file__).resolve().parents[1] / "generated_test_pdfs"


def get_default_manifest_path(output_dir: str | Path) -> Path:
    return Path(output_dir) / "manifest.json"


def load_evidence_fixture(path: str | Path) -> list[dict[str, Any]]:
    data = json.loads(Path(path).read_text(encoding="utf-8"))
    evidences = data.get("evidences")
    if not isinstance(evidences, list):
        raise ValueError("Expected top-level 'evidences' array in fixture JSON.")
    return evidences


def _resource_entry(evidence_wrapper: dict[str, Any]) -> tuple[str, dict[str, Any], dict[str, Any]]:
    evidence = evidence_wrapper.get("evidence")
    if not isinstance(evidence, dict):
        raise ValueError("Each fixture item must contain an 'evidence' object.")

    resource = evidence.get("resource")
    if not isinstance(resource, dict) or len(resource) != 1:
        raise ValueError("Each evidence object must contain exactly one resource entry.")

    resource_type = next(iter(resource))
    resource_body = resource[resource_type]
    if not isinstance(resource_body, dict):
        raise ValueError("Resource entry must be a JSON object.")

    return resource_type, resource_body, evidence


def _join_list(values: Any) -> str:
    if not isinstance(values, list) or not values:
        return "None documented."
    return ", ".join(str(value) for value in values)


def _validated_by_text(value: Any) -> str:
    if not isinstance(value, dict):
        return "No validation metadata documented."
    format_name = value.get("format", "unknown")
    schema_url = value.get("schemaUrl", "unknown")
    return f"Validation format: {format_name}. Referenced schema: {schema_url}."


def _generic_sections(resource_type: str, resource_body: dict[str, Any]) -> list[str]:
    lines = [
        f"{resource_body.get('name', resource_type)}",
        "",
        resource_body.get("description", ""),
    ]
    for key in sorted(resource_body):
        if key in {"id", "name", "description", "raw", "labels", "creationTime"}:
            continue
        value = resource_body[key]
        if isinstance(value, dict):
            rendered = json.dumps(value, sort_keys=True)
        elif isinstance(value, list):
            rendered = _join_list(value)
        else:
            rendered = str(value)
        lines.append(f"{key}: {rendered}")
    return lines


def _document_title(resource_type: str, resource_body: dict[str, Any]) -> str:
    titles = {
        "product": f"{resource_body.get('name', 'Product')} Product Overview",
        "application": f"{resource_body.get('name', 'Application')} Architecture Description",
        "contactPerson": "Product Security Contact Information",
        "monitoringProcedure": f"{resource_body.get('name', 'Monitoring Procedure')} Plan",
        "memory": f"{resource_body.get('name', 'Memory Component')} Technical Sheet",
        "virtualMachine": f"{resource_body.get('name', 'Virtual Machine')} Infrastructure Description",
        "userInformationAndIntructionDocument": "User Information and Instruction Manual",
        "sbomDocument": "Software Bill of Materials",
        "euDeclarationOfConformity": "EU Declaration of Conformity",
        "distributionOfUpdatesDocument": "Security Update Distribution Procedure",
        "productionAndMonitoringProcessDocument": "Production and Monitoring Process Description",
        "cyberSecurityRiskAssessmentDocument": "Cyber Security Risk Assessment",
        "coordinatedVulnerabilityDisclosurePolicy": "Coordinated Vulnerability Disclosure Policy",
    }
    return titles.get(resource_type, resource_body.get("name", resource_type))


def _build_cra_document_sections(resource_type: str, resource_body: dict[str, Any]) -> list[str]:
    templates: dict[str, list[str]] = {
        "product": [
            resource_body["name"],
            "Product Overview",
            resource_body["description"],
            (
                f"The product is classified as {resource_body.get('type', 'a product with digital elements')} "
                f"and is intended to {resource_body.get('purpose', 'support its documented purpose')}."
            ),
            f"Context of use: {resource_body.get('contextOfUse', 'Not documented.')}",
            "",
            "CRA evidence package",
            "The product evidence package includes an application architecture description, a user information and instruction manual,",
            "a software bill of materials, an EU declaration of conformity, a distribution of updates description,",
            "a production and monitoring process description, a cyber security risk assessment, and a coordinated",
            "vulnerability disclosure policy. The product scope also references a monitoring procedure, a memory",
            "component, and a management virtual machine as supporting evidence.",
        ],
        "application": [
            resource_body["name"],
            "Application Description",
            resource_body["description"],
            (
                "This application is operated as the control-plane service for connected edge sensors and "
                "supports the management backend used by the product. It "
                f"runs on a documented compute environment and is implemented in {resource_body.get('programmingLanguage', 'the documented language')}."
            ),
            f"Version: {resource_body.get('programmingVersion', 'Not documented.')}",
        ],
        "contactPerson": [
            resource_body["name"],
            "Responsible Contact",
            resource_body["description"],
            f"Role: {resource_body.get('jobTitle', 'Not documented.')}",
            f"Email address: {resource_body.get('emailAddress', 'Not documented.')}",
            f"Phone number: {resource_body.get('phoneNumber', 'Not documented.')}",
        ],
        "monitoringProcedure": [
            resource_body["name"],
            "Monitoring Procedure",
            resource_body["description"],
            (
                f"This monitoring procedure is scheduled every "
                f"{resource_body.get('intervalMonths', 'unknown')} month(s)."
            ),
        ],
        "memory": [
            resource_body["name"],
            "Hardware Component Description",
            resource_body["description"],
            f"Memory mode: {resource_body.get('mode', 'Not documented.')}",
        ],
        "virtualMachine": [
            resource_body["name"],
            "Infrastructure Component",
            resource_body["description"],
            (
                "Internet accessible endpoint: "
                f"{'Yes' if resource_body.get('internetAccessibleEndpoint') else 'No'}"
            ),
        ],
        "userInformationAndIntructionDocument": [
            resource_body["name"],
            "User Information and Instructions",
            resource_body["description"],
            (
                "This document provides user guidance for installation, operation, and support of the "
                "product with digital elements."
            ),
            f"File type: {resource_body.get('filetype', 'application/pdf')}",
            _validated_by_text(resource_body.get("validatedBy")),
        ],
        "sbomDocument": [
            resource_body["name"],
            "Software Bill of Materials",
            resource_body["description"],
            (
                "This SBOM identifies software components, dependencies, and component metadata needed "
                "for CRA-related transparency and vulnerability handling."
            ),
            f"File type: {resource_body.get('filetype', 'application/pdf')}",
            _validated_by_text(resource_body.get("validatedBy")),
        ],
        "euDeclarationOfConformity": [
            resource_body["name"],
            "EU Declaration of Conformity",
            resource_body["description"],
            (
                "This declaration records the applicable directives, standards, and conformity statement "
                "for the product."
            ),
            f"File type: {resource_body.get('filetype', 'application/pdf')}",
            _validated_by_text(resource_body.get("validatedBy")),
        ],
        "distributionOfUpdatesDocument": [
            resource_body["name"],
            "Distribution of Updates",
            resource_body["description"],
            (
                "This document explains how security updates and patches are prepared, approved, and "
                "distributed to customers."
            ),
            f"File type: {resource_body.get('filetype', 'application/pdf')}",
            _validated_by_text(resource_body.get("validatedBy")),
        ],
        "productionAndMonitoringProcessDocument": [
            resource_body["name"],
            "Production and Monitoring Process",
            resource_body["description"],
            (
                "This process description covers production controls, quality assurance, and ongoing "
                "monitoring activities for the product."
            ),
            f"File type: {resource_body.get('filetype', 'application/pdf')}",
            _validated_by_text(resource_body.get("validatedBy")),
        ],
        "cyberSecurityRiskAssessmentDocument": [
            resource_body["name"],
            "Cyber Security Risk Assessment",
            resource_body["description"],
            (
                "This assessment documents threats, vulnerabilities, likelihood, impact, and mitigation "
                "strategies relevant to the CRA."
            ),
            f"File type: {resource_body.get('filetype', 'application/pdf')}",
            _validated_by_text(resource_body.get("validatedBy")),
        ],
        "coordinatedVulnerabilityDisclosurePolicy": [
            resource_body["name"],
            "Coordinated Vulnerability Disclosure Policy",
            resource_body["description"],
            (
                "This policy defines the intake, triage, coordination, and disclosure timelines for "
                "reported vulnerabilities."
            ),
        ],
    }
    return templates.get(resource_type, _generic_sections(resource_type, resource_body))


def build_pdf_text(resource_type: str, resource_body: dict[str, Any], evidence: dict[str, Any]) -> str:
    lines = [
        _document_title(resource_type, resource_body),
        "",
    ]
    lines.extend(_build_cra_document_sections(resource_type, resource_body))
    return "\n".join(lines)


def _wrap_text_lines(text: str, max_chars: int = 88) -> list[str]:
    wrapped: list[str] = []
    for raw_line in text.splitlines():
        line = raw_line.rstrip()
        if not line:
            wrapped.append("")
            continue

        remaining = line
        while len(remaining) > max_chars:
            split_at = remaining.rfind(" ", 0, max_chars + 1)
            if split_at <= 0:
                split_at = max_chars
            wrapped.append(remaining[:split_at].rstrip())
            remaining = remaining[split_at:].lstrip()
        wrapped.append(remaining)
    return wrapped


def write_pdf(path: Path, text: str) -> None:
    wrapped_lines = _wrap_text_lines(text)
    max_lines_per_page = 40
    pages = [
        wrapped_lines[index : index + max_lines_per_page]
        for index in range(0, len(wrapped_lines), max_lines_per_page)
    ] or [[]]

    objects: list[bytes] = []
    page_object_numbers: list[int] = []

    def add_object(payload: bytes) -> int:
        objects.append(payload)
        return len(objects)

    catalog_object_number = add_object(b"<< /Type /Catalog /Pages 0 0 R >>")
    pages_object_number = add_object(b"<< /Type /Pages /Kids [] /Count 0 >>")
    font_object_number = add_object(b"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")

    for page_lines in pages:
        content_stream = _build_pdf_content_stream(page_lines)
        content_object_number = add_object(
            b"<< /Length "
            + str(len(content_stream)).encode("ascii")
            + b" >>\nstream\n"
            + content_stream
            + b"\nendstream"
        )
        page_object_number = add_object(
            (
                "<< /Type /Page /Parent "
                f"{pages_object_number} 0 R "
                "/MediaBox [0 0 612 792] "
                f"/Resources << /Font << /F1 {font_object_number} 0 R >> >> "
                f"/Contents {content_object_number} 0 R >>"
            ).encode("ascii")
        )
        page_object_numbers.append(page_object_number)

    objects[catalog_object_number - 1] = (
        f"<< /Type /Catalog /Pages {pages_object_number} 0 R >>".encode("ascii")
    )
    kids = " ".join(f"{number} 0 R" for number in page_object_numbers)
    objects[pages_object_number - 1] = (
        f"<< /Type /Pages /Kids [{kids}] /Count {len(page_object_numbers)} >>".encode("ascii")
    )

    pdf = bytearray(b"%PDF-1.4\n")
    offsets = [0]
    for index, obj in enumerate(objects, start=1):
        offsets.append(len(pdf))
        pdf.extend(f"{index} 0 obj\n".encode("ascii"))
        pdf.extend(obj)
        pdf.extend(b"\nendobj\n")

    xref_offset = len(pdf)
    pdf.extend(f"xref\n0 {len(objects) + 1}\n".encode("ascii"))
    pdf.extend(b"0000000000 65535 f \n")
    for offset in offsets[1:]:
        pdf.extend(f"{offset:010d} 00000 n \n".encode("ascii"))

    pdf.extend(
        f"trailer\n<< /Size {len(objects) + 1} /Root {catalog_object_number} 0 R >>\n".encode(
            "ascii"
        )
    )
    pdf.extend(f"startxref\n{xref_offset}\n%%EOF\n".encode("ascii"))
    path.write_bytes(pdf)


def _build_pdf_content_stream(lines: Sequence[str]) -> bytes:
    commands = [
        "BT",
        "/F1 10 Tf",
        "72 740 Td",
    ]
    for index, line in enumerate(lines):
        escaped = (
            line.replace("\\", "\\\\")
            .replace("(", "\\(")
            .replace(")", "\\)")
        )
        if index:
            commands.append("0 -16 Td")
        commands.append(f"({escaped}) Tj")
    commands.append("ET")
    return "\n".join(commands).encode("latin-1", "replace")


def generate_fixture_pdfs(
    input_path: str | Path,
    output_dir: str | Path,
) -> list[Path]:
    selected_resource_types = set(DEFAULT_RESOURCE_TYPES)
    output_root = Path(output_dir)
    output_root.mkdir(parents=True, exist_ok=True)

    generated_paths: list[Path] = []
    manifest_entries: list[dict[str, str]] = []
    fixture_index = 0
    for evidence_wrapper in load_evidence_fixture(input_path):
        resource_type, resource_body, evidence = _resource_entry(evidence_wrapper)
        if resource_type not in selected_resource_types:
            continue

        fixture_index += 1
        output_path = output_root / f"cra-fixture-{fixture_index:02d}.pdf"
        write_pdf(output_path, build_pdf_text(resource_type, resource_body, evidence))
        generated_paths.append(output_path)
        manifest_entries.append(
            {
                "filename": output_path.name,
                "expectedResourceType": resource_type,
                "resourceId": str(resource_body.get("id", "")),
                "resourceName": str(resource_body.get("name", "")),
                "description": str(resource_body.get("description", "")),
            }
        )

    get_default_manifest_path(output_root).write_text(
        json.dumps({"fixtures": manifest_entries}, indent=2),
        encoding="utf-8",
    )

    return generated_paths


def main(argv: Sequence[str] | None = None) -> int:
    args = parse_args(argv)
    generated_paths = generate_fixture_pdfs(args.input_path, args.output_dir)
    print(f"Generated {len(generated_paths)} PDF fixture(s) in {Path(args.output_dir).resolve()}")
    for path in generated_paths:
        print(path.resolve())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
