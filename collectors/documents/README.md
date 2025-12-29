# Document-Analyser

Turns unstructured documents (PDF or text) into structured evidence items with snippets and citations. The pipeline is intentionally small:
- `loaders.DocumentLoader` reads files and normalizes their content.
- `prompts` + `extractor.DocumentAnalyser` build the prompt and parse the LLM JSON output.
- `pipeline.DocumentAnalysisPipeline` orchestrates load → extract → optional push.
- `evidence_store.EvidencePublisher` formats evidence and sends it to an OAuth-protected evidence store.
- `requirements` holds predefined requirement prompts you can run individually or all at once.

## Quick start
1) Install
```bash
python -m venv .venv
source .venv/bin/activate
pip install -e .
```

2) Provide model access  
Set `OPENAI_API_KEY` (or `DOC_ANALYSER_API_KEY`) for remote models. For local OpenAI-compatible servers, set `DOC_ANALYSER_BASE_URL` and any placeholder `DOC_ANALYSER_API_KEY`, then choose the model with `--model`.

3) Run the CLI
- General analysis:  
  `python -m document_analyser.cli file.pdf --model gpt-4o-mini --focus "encryption controls"`
- Local model:  
  `python -m document_analyser.cli file.pdf --model llama3 --base-url http://localhost:11434/v1 --api-key local`
- Single requirement:  
  `python -m document_analyser.cli file.pdf --test-requirement X.1.1.6`
- All requirements:  
  `python -m document_analyser.cli file.pdf --all-requirements`
- Push to evidence store:  
  `python -m document_analyser.cli file.pdf --push-evidence --evidence-url http://localhost:8080 --evidence-auth client:secret --evidence-target-id <uuid> --evidence-tool-id document-analyser`
- Evidence-store connectivity test only:  
  `python -m document_analyser.cli --test-evidence --evidence-url http://localhost:8080 --evidence-auth client:secret`

Copy-paste examples:
```bash
# 1) Remote model quick run
python -m document_analyser.cli policy.pdf --model gpt-4o-mini --focus "access control"

# 2) Local model (OpenAI-compatible endpoint)
python -m document_analyser.cli file.pdf --model llama4:maverick --base-url http://172.31.6.131:11434/v1 --api-key local

# 3) Push evidence after analysing all requirements
python -m document_analyser.cli file.pdf --all-requirements --push-evidence \
  --evidence-url http://localhost:8080 --evidence-auth clouditor:clouditor --evidence-target-id 00000000-0000-0000-0000-000000000000
```

## CLI flags (trimmed)
- LLM: `--model`, `--base-url`, `--api-key`, `--focus`, `--max-items`
- Evidence store: `--push-evidence`, `--evidence-url`, `--evidence-auth client:secret`, `--evidence-target-id`, `--evidence-tool-id`
- Modes: `--test-requirement <ID>`, `--all-requirements`, `--test-evidence`

Env defaults for the evidence store:  
`EVIDENCE_STORE_BASE` (or `CONFIRMATE_API_BASE`, default `http://localhost:8080`), `AUTH_TOKEN_ENDPOINT` (defaults to `<base>/v1/auth/token`), `AUTH_CLIENT_ID`, `AUTH_CLIENT_SECRET`, `TARGET_OF_EVALUATION_ID`, `EVIDENCE_TOOL_ID`, `EVIDENCE_PATH` (default `/v1/evidence_store/evidence`). `EVIDENCE_AUTO_PUSH=true` auto-pushes after analysis.

## How it works
1) Loader: `loaders.py` reads PDF (via pypdf) or text, joins multiple files with headers.  
2) Prompting: `prompts.py` builds the system/user messages for general extraction or per-requirement extraction.  
3) LLM call: `llm.py` wraps the OpenAI-compatible chat client, requesting JSON output.  
4) Parsing: `extractor.py` runs the LLM, JSON-parses the reply, and packages an `AnalysisResult`. It can run the general extractor or one evidence item per requirement.  
5) Pipeline: `pipeline.py` stitches loader + extractor and optionally a publisher for a single `run()` entrypoint.  
6) Publishing: `evidence_store.py` converts evidence items to the target schema and sends them with OAuth.  
7) CLI: `cli.py` wires flags → config → pipeline.  
8) Requirements: `requirements.py` stores predefined requirement prompts and IDs.

## File map
- `src/document_analyser/cli.py` – command-line entrypoint.
- `src/document_analyser/config.py` – model config from env and overrides.
- `src/document_analyser/loaders.py` – file loading and concatenation.
- `src/document_analyser/prompts.py` – prompt templates for general/requirement runs.
- `src/document_analyser/llm.py` – OpenAI-compatible client wrapper.
- `src/document_analyser/extractor.py` – LLM orchestration and JSON parsing.
- `src/document_analyser/pipeline.py` – load → extract → optional push coordinator.
- `src/document_analyser/evidence_store.py` – payload building, OAuth, submission.
- `src/document_analyser/requirements.py` – predefined requirement prompts.

## Example output (general analysis)
```json
{
  "sources": ["policy.md"],
  "analysis": {
    "document_summary": "Access control policy describing authentication, MFA and audit logging.",
    "evidence": [
      {
        "title": "Authentication",
        "evidence": "Passwords must be rotated every 90 days with MFA enforced for admins.",
        "snippet": "Section 2.1 'Authentication'",
        "citation": "Section 2.1 'Authentication'",
        "confidence": "high"
      }
    ],
    "gaps": []
  },
  "raw_response": "{...raw model output...}"
}
```
