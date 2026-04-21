# Document-Analyser

Document-Analyser turns PDFs and text documents into structured evidence for Confirmate.

It supports two separate analysis modes:

1. `requirements`
Checks the predefined CRA requirements from `requirements.py`.

2. `resources`
Extracts ontology-based resource evidence using the CRA-focused whitelist by default.

It can also:
- read a single file or an entire directory
- use remote OpenAI-compatible models or local LLM endpoints
- push extracted evidence to the Evidence Store
- generate a PDF test dataset for CRA-relevant resource types

## What It Does

The pipeline is intentionally small:

- `loaders.DocumentLoader` reads PDF or text files and normalizes their content
- `prompts` + `extractor.DocumentAnalyser` build prompts and parse JSON output from the LLM
- `pipeline.DocumentAnalysisPipeline` orchestrates load -> extract -> optional push
- `evidence_store.EvidencePublisher` converts extracted results into Evidence Store payloads
- `requirements` stores the predefined CRA requirement prompts
- `evidence_profiles` and `proto_schema` derive supported ontology resource types and fields from the ontology proto

## Supported Modes

### 1. CRA Requirement Checks

Mode: `--mode requirements`

This mode checks documents against the predefined CRA requirements in `src/document_analyser/requirements.py`.

Use this when you want to know whether the document contains evidence for CRA requirement statements such as confidentiality protection, access control, security updates, logging, and similar checks.

Example:
```bash
python -m document_analyser.cli file.pdf --mode requirements
```

Single requirement:
```bash
python -m document_analyser.cli file.pdf --mode requirements --test-requirement X.1.1.6
```

### 2. Ontology Whitelist Resource Extraction

Mode: `--mode resources`

This mode extracts ontology-backed resource evidence from documents.

By default it uses the CRA-focused whitelist of resource types:
- `product`
- `application`
- `contactPerson`
- `monitoringProcedure`
- `memory`
- `virtualMachine`
- `userInformationAndIntructionDocument`
- `sbomDocument`
- `euDeclarationOfConformity`
- `distributionOfUpdatesDocument`
- `productionAndMonitoringProcessDocument`
- `cyberSecurityRiskAssessmentDocument`
- `coordinatedVulnerabilityDisclosurePolicy`

Example:
```bash
python -m document_analyser.cli file.pdf --mode resources
```

If you want to allow all ontology resource types derived from the proto schema:
```bash
python -m document_analyser.cli file.pdf --mode resources --all-resource-types
```

## Input Options

You can pass either:

- a single file
- multiple files
- a directory

When a directory is passed, the CLI recursively collects supported document files inside it.

Example:
```bash
python -m document_analyser.cli generated_test_pdfs --mode resources
```

## Local LLM Usage

If you are using a local OpenAI-compatible model endpoint, set:

- `DOC_ANALYSER_BASE_URL`
- `DOC_ANALYSER_API_KEY`

Example:
```bash
DOC_ANALYSER_BASE_URL=http://172.31.6.131:8000/v1 \
DOC_ANALYSER_API_KEY=local \
PYTHONPATH=src \
python3 -m document_analyser.cli generated_test_pdfs \
  --model Qwen/Qwen3.5-122B-A10B-FP8 \
  --mode resources
```

Another available model example:
```bash
DOC_ANALYSER_BASE_URL=http://172.31.6.131:8000/v1 \
DOC_ANALYSER_API_KEY=local \
PYTHONPATH=src \
python3 -m document_analyser.cli generated_test_pdfs \
  --model SASAI \
  --mode resources
```

## Evidence Store Usage

To push extracted evidence to the Evidence Store, use:

- `--push-evidence`
- `--evidence-url`
- `--evidence-token`

If your bearer token is already stored in `$TOKEN`:
```bash
DOC_ANALYSER_BASE_URL=http://172.31.6.131:8000/v1 \
DOC_ANALYSER_API_KEY=local \
PYTHONPATH=src \
python3 -m document_analyser.cli generated_test_pdfs \
  --model Qwen/Qwen3.5-122B-A10B-FP8 \
  --mode resources \
  --push-evidence \
  --evidence-url http://localhost:8080 \
  --evidence-token "$TOKEN"
```

CRA requirement mode with Evidence Store push:
```bash
DOC_ANALYSER_BASE_URL=http://172.31.6.131:8000/v1 \
DOC_ANALYSER_API_KEY=local \
PYTHONPATH=src \
python3 -m document_analyser.cli generated_test_pdfs \
  --model Qwen/Qwen3.5-122B-A10B-FP8 \
  --mode requirements \
  --push-evidence \
  --evidence-url http://localhost:8080 \
  --evidence-token "$TOKEN"
```

## Test Dataset Generation

The repository includes a script that generates a CRA-focused PDF test dataset.

Script:
```bash
python3 scripts/generate_test_pdfs.py
```

Example:
```bash
cd collectors/documents
PYTHONPATH=src python3 scripts/generate_test_pdfs.py
```

This creates:
- `generated_test_pdfs/cra-fixture-01.pdf`, `cra-fixture-02.pdf`, etc.
- `generated_test_pdfs/manifest.json`

The PDFs are neutral test inputs.
The `manifest.json` file contains the expected resource type for each PDF, so you can compare analyser output against the ground truth.

## Recommended Testing Workflow

### Test CRA requirement mode
```bash
DOC_ANALYSER_BASE_URL=http://172.31.6.131:8000/v1 \
DOC_ANALYSER_API_KEY=local \
PYTHONPATH=src \
python3 -m document_analyser.cli generated_test_pdfs \
  --model Qwen/Qwen3.5-122B-A10B-FP8 \
  --mode requirements
```

### Test ontology whitelist resource mode
```bash
DOC_ANALYSER_BASE_URL=http://172.31.6.131:8000/v1 \
DOC_ANALYSER_API_KEY=local \
PYTHONPATH=src \
python3 -m document_analyser.cli generated_test_pdfs \
  --model Qwen/Qwen3.5-122B-A10B-FP8 \
  --mode resources
```

### Test one file at a time
```bash
DOC_ANALYSER_BASE_URL=http://172.31.6.131:8000/v1 \
DOC_ANALYSER_API_KEY=local \
PYTHONPATH=src \
python3 -m document_analyser.cli generated_test_pdfs/cra-fixture-01.pdf \
  --model Qwen/Qwen3.5-122B-A10B-FP8 \
  --mode resources
```

## CLI Options

### Core
- `files`
  One or more files or directories to analyse
- `--mode {requirements,resources}`
  Choose CRA requirement checks or ontology whitelist resource extraction
- `--model`
  Model name
- `--base-url`
  OpenAI-compatible base URL
- `--api-key`
  API key or placeholder for local endpoints
- `--focus`
  Optional focus hint for general analysis
- `--max-items`
  Maximum number of returned evidence items

### Requirement Mode
- `--test-requirement <ID>`
  Run a single CRA requirement

### Resource Mode
- `--all-resource-types`
  Use all proto-derived ontology resource types instead of the default CRA whitelist

### Evidence Store
- `--push-evidence`
  Send generated evidence to the Evidence Store
- `--evidence-url`
  Evidence Store base URL
- `--evidence-auth`
  Client credentials in `client_id:client_secret` format
- `--evidence-token`
  Bearer token, e.g. `"$TOKEN"`
- `--evidence-target-id`
  Override target of evaluation ID
- `--evidence-tool-id`
  Override tool ID

### Utility Modes
- `--test-evidence`
  Send a minimal connectivity test payload to the Evidence Store
- `--push-evidence-file <path>`
  Push prebuilt evidence payloads from a JSON file

## File Map

- `src/document_analyser/cli.py`
  Command-line entrypoint
- `src/document_analyser/config.py`
  Model configuration
- `src/document_analyser/loaders.py`
  File loading and PDF extraction
- `src/document_analyser/prompts.py`
  Prompt construction
- `src/document_analyser/extractor.py`
  LLM orchestration and JSON parsing
- `src/document_analyser/pipeline.py`
  Load -> extract -> optional push
- `src/document_analyser/evidence_store.py`
  Evidence Store payload building and submission
- `src/document_analyser/evidence_profiles.py`
  Proto-derived ontology resource profiles
- `src/document_analyser/proto_schema.py`
  Resource normalization helpers
- `src/document_analyser/requirements.py`
  CRA requirement prompt definitions
- `scripts/generate_test_pdfs.py`
  CRA PDF test dataset generator

## Notes

- `requirements` mode and `resources` mode are intentionally separate
- `resources` mode uses the CRA-focused whitelist by default
- passing a directory is supported
- for local LLMs, using `DOC_ANALYSER_BASE_URL` and `DOC_ANALYSER_API_KEY=local` is the most reliable setup
