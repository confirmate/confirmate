from __future__ import annotations

import re
from dataclasses import dataclass
from functools import lru_cache
from pathlib import Path
from typing import Dict, Iterable


SCALAR_FIELD_TYPES = {
    "bool",
    "bytes",
    "double",
    "fixed32",
    "fixed64",
    "float",
    "int32",
    "int64",
    "sfixed32",
    "sfixed64",
    "sint32",
    "sint64",
    "string",
    "uint32",
    "uint64",
}

RESOURCE_CONTAINERS = ("Resource", "Data", "Policy")
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


@dataclass(frozen=True)
class ProtoField:
    name: str
    proto_name: str
    type_name: str
    required: bool
    repeated: bool = False
    optional: bool = False
    map_field: bool = False
    description: str = ""

    @property
    def prompt_type(self) -> str:
        if self.map_field:
            return "map"
        if self.type_name in SCALAR_FIELD_TYPES:
            return self.type_name
        if self.type_name.startswith("google.protobuf."):
            return self.type_name.rsplit(".", 1)[-1]
        return self.type_name


@dataclass(frozen=True)
class EvidenceResourceProfile:
    resource_type: str
    message_name: str
    description: str
    categories: tuple[str, ...]
    fields: tuple[ProtoField, ...]
    required_fields: tuple[str, ...]
    optional_fields: tuple[str, ...]
    supports_data_location: bool
    supports_filetype: bool


@dataclass(frozen=True)
class ProtoMessageDefinition:
    name: str
    description: str
    resource_type_names: tuple[str, ...]
    fields: tuple[ProtoField, ...]


def _snake_to_camel(value: str) -> str:
    parts = [part for part in value.split("_") if part]
    if not parts:
        return value
    return parts[0] + "".join(part[:1].upper() + part[1:] for part in parts[1:])


def _pascal_to_camel(value: str) -> str:
    parts = re.findall(r"[A-Z]+(?=[A-Z][a-z0-9]|$)|[A-Z]?[a-z0-9]+", value)
    if not parts:
        return value[:1].lower() + value[1:]
    return parts[0].lower() + "".join(part[:1].upper() + part[1:] for part in parts[1:])


def _normalize_token(value: str) -> str:
    return "".join(ch for ch in value.lower() if ch.isalnum())


def _extract_comment_block(text: str) -> str:
    lines = []
    for line in text.splitlines():
        cleaned = line.strip()
        if not cleaned.startswith("//"):
            continue
        lines.append(cleaned[2:].strip())
    return " ".join(line for line in lines if line)


def _parse_field_line(line: str, description: str = "") -> ProtoField | None:
    pattern = re.compile(
        r"^(?:(?P<label>repeated|optional)\s+)?"
        r"(?P<type>map<[^>]+>|[\w.]+)\s+"
        r"(?P<name>\w+)\s*=\s*\d+"
        r"(?P<attrs>\s*\[[^\]]+\])?\s*;$"
    )
    match = pattern.match(line.strip())
    if not match:
        return None

    proto_name = match.group("name")
    type_name = match.group("type")
    label = match.group("label") or ""
    attrs = match.group("attrs") or ""
    return ProtoField(
        name=_snake_to_camel(proto_name),
        proto_name=proto_name,
        type_name=type_name,
        required=".required = true" in attrs,
        repeated=label == "repeated",
        optional=label == "optional",
        map_field=type_name.startswith("map<"),
        description=description,
    )


def _parse_message_fields(block: str) -> tuple[ProtoField, ...]:
    fields = []
    pending_comments = []
    brace_level = 0

    for raw_line in block.splitlines():
        line = raw_line.strip()
        if not line:
            pending_comments = []
            continue

        if line.startswith("//"):
            pending_comments.append(line)
            continue

        if line.startswith("oneof "):
            brace_level += raw_line.count("{") - raw_line.count("}")
            pending_comments = []
            continue

        if brace_level > 0:
            brace_level += raw_line.count("{") - raw_line.count("}")
            if brace_level < 0:
                brace_level = 0
            pending_comments = []
            continue

        field = _parse_field_line(line, description=_extract_comment_block("\n".join(pending_comments)))
        pending_comments = []
        if field is not None:
            fields.append(field)

    return tuple(fields)


def _find_matching_brace(text: str, start_index: int) -> int:
    depth = 0
    for index in range(start_index, len(text)):
        char = text[index]
        if char == "{":
            depth += 1
        elif char == "}":
            depth -= 1
            if depth == 0:
                return index
    raise ValueError("Could not find matching closing brace in ontology.proto")


def _get_repo_root() -> Path:
    return Path(__file__).resolve().parents[4]


def get_ontology_proto_path() -> Path:
    return _get_repo_root() / "core" / "policies" / "security-metrics" / "ontology" / "v1" / "ontology.proto"


@lru_cache(maxsize=1)
def load_proto_message_definitions() -> Dict[str, ProtoMessageDefinition]:
    text = get_ontology_proto_path().read_text(encoding="utf-8")
    pattern = re.compile(r"(?P<comments>(?:(?:\s*//[^\n]*\n)*))\s*message\s+(?P<name>\w+)\s*\{")

    definitions: Dict[str, ProtoMessageDefinition] = {}
    for match in pattern.finditer(text):
        name = match.group("name")
        block_start = match.end() - 1
        block_end = _find_matching_brace(text, block_start)
        block = text[block_start + 1 : block_end]
        resource_type_names = tuple(
            re.findall(r'option\s+\(resource_type_names\)\s*=\s*"([^"]+)";', block)
        )
        definitions[name] = ProtoMessageDefinition(
            name=name,
            description=_extract_comment_block(match.group("comments")),
            resource_type_names=resource_type_names,
            fields=_parse_message_fields(block),
        )

    return definitions


@lru_cache(maxsize=1)
def _load_container_entries() -> Dict[str, tuple[tuple[str, str], ...]]:
    definitions = load_proto_message_definitions()
    entries: Dict[str, tuple[tuple[str, str], ...]] = {}

    for container in RESOURCE_CONTAINERS:
        definition = definitions.get(container)
        if definition is None:
            continue

        container_entries = []
        text = get_ontology_proto_path().read_text(encoding="utf-8")
        container_marker = f"message {container} {{"
        start = text.find(container_marker)
        if start < 0:
            continue
        block_start = text.find("{", start)
        block_end = _find_matching_brace(text, block_start)
        block = text[block_start + 1 : block_end]
        oneof_match = re.search(r"oneof\s+type\s*\{(?P<body>.*?)\n\s*\}", block, re.S)
        if not oneof_match:
            entries[container] = tuple()
            continue

        for raw_line in oneof_match.group("body").splitlines():
            line = raw_line.strip()
            field = _parse_field_line(line)
            if field is None:
                continue
            container_entries.append((field.type_name, _snake_to_camel(field.proto_name)))

        entries[container] = tuple(container_entries)

    return entries


@lru_cache(maxsize=1)
def get_resource_profiles() -> Dict[str, EvidenceResourceProfile]:
    definitions = load_proto_message_definitions()
    container_entries = _load_container_entries()

    profile_builders: Dict[str, dict[str, object]] = {}
    for container, entries in container_entries.items():
        for message_name, resource_type in entries:
            message_definition = definitions.get(message_name)
            if message_definition is None:
                continue

            builder = profile_builders.setdefault(
                resource_type,
                {
                    "message_name": message_name,
                    "description": message_definition.description,
                    "categories": set(),
                    "fields": message_definition.fields,
                },
            )
            builder["categories"].add(container)
            for category in message_definition.resource_type_names:
                builder["categories"].add(category)

    profiles: Dict[str, EvidenceResourceProfile] = {}
    for resource_type, builder in sorted(profile_builders.items()):
        fields = tuple(builder["fields"])
        required_fields = tuple(field.name for field in fields if field.required)
        optional_fields = tuple(field.name for field in fields if not field.required)
        profiles[resource_type] = EvidenceResourceProfile(
            resource_type=resource_type,
            message_name=str(builder["message_name"]),
            description=str(builder["description"]),
            categories=tuple(sorted(str(category) for category in builder["categories"] if category)),
            fields=fields,
            required_fields=required_fields,
            optional_fields=optional_fields,
            supports_data_location=any(field.name == "dataLocation" for field in fields),
            supports_filetype=any(field.name == "filetype" for field in fields),
        )

    return profiles


@lru_cache(maxsize=1)
def get_resource_type_aliases() -> Dict[str, str]:
    definitions = load_proto_message_definitions()
    aliases: Dict[str, str] = {}

    for resource_type, profile in get_resource_profiles().items():
        aliases[_normalize_token(resource_type)] = resource_type
        aliases[_normalize_token(profile.message_name)] = resource_type

        definition = definitions.get(profile.message_name)
        if definition is not None:
            for alias in definition.resource_type_names:
                aliases[_normalize_token(alias)] = resource_type
                aliases[_normalize_token(_pascal_to_camel(alias))] = resource_type

    return aliases


def normalize_resource_type(value: str | None) -> str | None:
    if not value:
        return None
    return get_resource_type_aliases().get(_normalize_token(value))


def iter_resource_profiles() -> Iterable[EvidenceResourceProfile]:
    return get_resource_profiles().values()


def get_default_resource_types() -> tuple[str, ...]:
    profiles = get_resource_profiles()
    return tuple(resource_type for resource_type in DEFAULT_RESOURCE_TYPES if resource_type in profiles)


def get_selected_resource_profiles(
    include_all: bool = False,
) -> Dict[str, EvidenceResourceProfile]:
    profiles = get_resource_profiles()
    if include_all:
        return profiles

    selected = get_default_resource_types()
    return {resource_type: profiles[resource_type] for resource_type in selected}


def render_resource_profiles_for_prompt(include_all: bool = False) -> str:
    lines = []
    for profile in get_selected_resource_profiles(include_all=include_all).values():
        field_descriptions = []
        for field in profile.fields:
            modifiers = []
            if field.required:
                modifiers.append("required")
            if field.repeated:
                modifiers.append("repeated")
            elif field.optional:
                modifiers.append("optional")
            modifier_text = f" [{', '.join(modifiers)}]" if modifiers else ""
            field_descriptions.append(f"{field.name}:{field.prompt_type}{modifier_text}")

        category_text = ", ".join(profile.categories) if profile.categories else "uncategorized"
        lines.append(
            f'- "{profile.resource_type}" ({profile.message_name}; categories: {category_text}): '
            f"{profile.description or 'Ontology resource extracted from the proto schema.'} "
            f"Fields: {', '.join(field_descriptions) if field_descriptions else 'none'}."
        )
    return "\n".join(lines)
