# CRA compliance management: product design

Status: draft, not implemented. Branch `docs/cra-product-design`, no PR yet.

References to the Cyber Resilience Act (Regulation (EU) 2024/2847, "CRA") are from our reading of the text and have to
be reviewed by someone with legal and compliance expertise before the catalog content is released (see
[Catalog](#5-cra-catalog)).

## 1. Purpose and scope

A company that places a software product on the EU market can manage its CRA compliance end to end in Confirmate:

1. describe the product and learn which obligations apply,
2. connect the systems that already hold the evidence (git platform, Dependency-Track, issue tracker, documents),
3. see continuously what is satisfied, what is missing and what is overdue,
4. work off the open items (tasks, documents, manual statements, deadlines),
5. produce the technical file and the EU declaration of conformity, and keep proof of how they came to be.

It has to be usable by a compliance manager who does not know the Confirmate data model and by an engineer who does not
know the CRA text.

### Non-goals

- No SBOM generation, vulnerability matching or triage UI. Dependency-Track (DT) does this; we link to it.
- No ticket system. Tasks in Confirmate exist for compliance work only; issues stay in GitHub/Jira.
- No legal advice. The catalog describes what evidence supports which requirement; the legal conclusion stays with the
  manufacturer.
- Hardware products and cloud-only services outside the CRA scope are not a target. (The cloud collectors remain useful
  for the infrastructure a product runs on.)
- Multi-tenancy (several unrelated companies in one installation) is not part of the first versions.

## 2. Decisions

| # | Decision | Reason |
|---|----------|--------|
| D1 | The product is the central object; it maps 1:1 to a target of evaluation (ToE) and its CRA audit scope. | Reuses orchestrator, evaluation, permissions and reports unchanged. |
| D2 | One organization per installation. Several products per organization. | Smallest model that serves the pilots; permissions per product already exist. |
| D3 | DT is the source of truth for SBOMs, components, vulnerabilities and their analysis state (VEX). Confirmate reads from DT and never writes to it. | We do not maintain an SBOM solution; read-only keeps the trust boundary simple. |
| D4 | Target DT v5 (`DependencyTrack/dependency-track` `main`). v4 is out of scope. | v5 is GA and the future; v4 is in maintenance. |
| D5 | CSAF in DT is an optional capability, detected at runtime. | It is not in a DT release yet. |
| D6 | Connections to external systems are configured through the Confirmate API/UI and run server-side. No flags/env vars for users. | Required for a guided onboarding; credentials are stored once, encrypted. |
| D7 | The CRA catalog is maintained in the `security-metrics` repository next to the ontology and metrics, versioned, and reviewed against the regulation. | One place for compliance content; the repo already ships catalogs. |
| D8 | Automatic evidence wherever possible; everything else is an explicit manual control with a defined evidence type (document, statement). No control without either. | Prevents "silent gaps" in the catalog. |
| D9 | Documents and statements without automatic evidence are authored or uploaded in Confirmate and stored as evidence with versions and approval. | The user manages compliance in the tool, not next to it. |
| D10 | Evidence history is append-only and snapshots are first class (state at a release or a date). | Audits ask "what was true when you shipped 2.3". |

## 3. CRA obligations and how they are served

This is the working mapping that drives the catalog. "Auto" means a collector provides the evidence, "doc" means a
document is the evidence, "manual" means a reviewed statement.

| Obligation (CRA) | Evidence | Source |
|------------------|----------|--------|
| Annex I Part I (1): appropriate level of cybersecurity based on the risks | Cyber security risk assessment document (Art. 13(2)) | doc |
| Part I (2)(a) no known exploitable vulnerabilities | Findings without exploitable state | auto (DT) |
| (b) secure by default, reset option | Configuration facts, documentation | auto (code analysis) + doc |
| (c) security updates, automatic updates by default with opt-out | Update mechanism facts | auto (code analysis) + doc |
| (d) protection from unauthorized access, reporting of access | Auth mechanisms, MFA | auto (code analysis, cloud) |
| (e) confidentiality (encryption at rest and in transit) | Encryption facts | auto (code analysis, cloud) |
| (f) integrity | Integrity protection facts | auto |
| (g) data minimisation | Data handling facts | auto (code analysis) + manual |
| (h)/(i) availability and impact on other services | Resilience facts | auto (cloud) + manual |
| (j) limit attack surfaces | Interface inventory | auto + manual |
| (k) reduce impact of incidents | Mitigation facts | manual |
| (l) logging and monitoring, opt-out | Logging facts | auto (cloud, code analysis) |
| (m) secure removal of data and settings | Deletion mechanism facts | auto + manual |
| Part II (1) identify components, SBOM in a common machine-readable format | SBOM exists, format, required fields, per release | auto (DT) |
| (2) fix vulnerabilities without delay; updates separate from features where feasible | Time to triage/fix against SLA | auto (DT, issue tracker) |
| (3) regular tests and reviews | Test and scan runs in CI | auto (git platform) |
| (4) publish fixed vulnerabilities after the update is available | Advisory exists for fixed findings | auto (DT CSAF, later), doc |
| (5)/(6) coordinated vulnerability disclosure policy, contact address | Policy document, `SECURITY.md`, reachable contact | doc + auto (git platform) |
| (7) secure distribution of updates | Signing and distribution facts | auto (later) + doc |
| (8) security patches without delay and free of charge, with advisories | Patch release timeline, advisories | auto + manual |
| Art. 13: due diligence for components | Component list with vulnerability state | auto (DT) |
| Art. 13: support period (at least five years unless the product is expected to be used for less) | Product attribute with end date and reminders | manual |
| Art. 13: documentation retained (ten years or the support period if longer) | Snapshots and retention rule | system |
| Art. 14: reporting of actively exploited vulnerabilities and severe incidents (early warning within 24 h, notification within 72 h, final report) | Clocks, drafts, submission proof | workflow |
| Annex VII technical documentation | Technical file generated from current evidence | output |
| Annex V EU declaration of conformity | Generated, approved, versioned | output |
| Annex VIII conformity route (internal control, or third party depending on the class) | Product class decides | product attribute |

Reporting obligations of Art. 14 apply earlier than the main obligations (to our knowledge from 11 September 2026 versus
11 December 2027; to be verified).

## 4. Domain model

New concepts, in addition to the existing ToE, catalog, control, metric, evidence, assessment and evaluation result.

```
Product            id, name, description, class (DEFAULT | IMPORTANT_I | IMPORTANT_II | CRITICAL),
                   support_end, target_of_evaluation_id, audit_scope_id, owner
Release            id, product_id, version, released_at, artifacts[], sbom_ref, snapshot_id
Connection         id, product_id, type (GITHUB | DEPENDENCY_TRACK | ISSUES_GITHUB | ISSUES_JIRA | ...),
                   config (non-secret), secret_ref, status, last_sync_at, last_error
Task               id, product_id, control_id?, title, assignee, due_at, status, source (SYSTEM | USER)
Deadline/Clock     id, product_id, kind (EARLY_WARNING_24H | NOTIFICATION_72H | FINAL_REPORT | SUPPORT_END | ...),
                   subject_ref, starts_at, due_at, state
Report             id, product_id, kind (VULNERABILITY | INCIDENT), stage, content, submitted_at, proof_ref
Snapshot           id, product_id, taken_at, label, content_hash   // references result ids; never copies evidence
```

- `Product` is created together with its ToE and audit scope on the generated catalog; the existing audit-scope
  workflow (control states, manual results, audit trail) is reused.
- `Connection.secret_ref` points into an encrypted secret store of the installation. Secrets are never returned by the
  API.
- `Task` and `Deadline` are generated by the system from evaluation results (failing control, expiring manual
  statement, security issue beyond its SLA, upcoming support end) and can be created by users. They are what the
  work queue shows.
- Snapshots reference immutable result ids and carry a content hash, so a technical file can be verified later.

API: standard resource-oriented methods (`CreateProduct`, `GetProduct`, `ListProducts`, `UpdateProduct`,
`CreateConnection`, `ListTasks`, ...) following the API guidelines in `CONTRIBUTING.md`.

## 5. CRA catalog

Replaces `example-data/catalogs/cra-example-catalog.json`.

- Covers Annex I Part I and II, Art. 13 and 14 obligations that can be evidenced, and Annex VII/V outputs.
- Each control has a CRA reference (article/annex/paragraph), a plain-language description, the expected evidence
  type and either metrics or a manual definition (D8).
- Product class metadata decides which controls are in scope (for example the conformity route controls).
- Category and control ids follow the CRA structure so that they stay stable across catalog versions.

What exists today, as the starting point: 13 parent controls `CRA-I-1.1.2` to `CRA-I-1.1.14`, which map one to one to
Annex I Part I (2)(a) to (m), and 27 leaf controls with metrics only on the leaves. Known quality problems to fix when
migrating: there is no control for Part I (1); some placement is questionable (for example "Password Policy
Enforcement" under integrity and "Critical Data Backup" under exploitation mitigation); 14 of the 26 referenced metrics
have no implementation in the metrics repository; the CSAF and vulnerability-count metrics exist but are not part of
the catalog; Part II, Art. 13/14 and Annex VII are missing entirely.

Review process: draft by us, structural review against the regulation text by a CRA-literate reviewer, then versioned
release with a changelog. Control wording stays neutral and states the evidence, not a legal conclusion.

## 6. Integrations

All integrations run in a server-side connector runtime (an extension of `core/service/collection`) that reads the
`Connection` configuration from the orchestrator. The existing collectors keep working standalone.

### Dependency-Track

Verified in the DT source (not yet against a running instance):

- the stable REST API `/api/v1` provides projects, BOM (CycloneDX), components, findings, VEX, policy violations and
  metrics (`ProjectResource`, `BomResource`, `FindingResource`, `VexResource`, `MetricsResource`);
- the newer spec-first API (`api/src/main/openapi`) covers only some resources so far and offers
  `/internal/system-capabilities`;
- CSAF (`/csaf/*`, `/advisories`) exists only on a fork branch.

Design:

- Per linked DT project the connector reads the latest BOM, components, findings with analysis state, VEX, policy
  violations and project metrics, and submits them as evidence.
- A DT webhook alert (BOM processed, new vulnerability) triggers an immediate refresh; polling is the fallback.
- The connector hides the API version behind an interface so the newer API can replace calls one by one.
- Advisories are an optional capability: when DT announces CSAF support the connector fills the advisory evidence,
  until then the related metrics report "not available".

Product to DT project matching: suggest candidates by name and package URL, user confirms (one click).

### Git platform (GitHub first, GitLab later)

Repository settings (branch protection, required reviews, signed commits, 2FA requirement, `SECURITY.md`, dependency
update bot), workflows and their recent runs, releases and their assets. Fills `CodeRepository`, `Workflow`, `Job`.
Connection by GitHub App installation.

### Issue tracker (GitHub issues first, then Jira)

Security-labelled issues with created, triaged and closed times, linked to DT findings where possible. Needs a new
ontology type. Used for the time-to-triage and time-to-fix metrics.

### Documents

The existing document analyser extracts evidence from uploaded documents. In addition the UI authors documents
directly (see [UX](#7-ux)); both produce the same `Document` evidence types.

### Ontology and metrics additions (to be proposed in `security-metrics`)

- `Vulnerability`: severity, analysis state, published/fixed/triaged timestamps, affected component.
- SBOM facts: format, spec version, component count, which required fields are present.
- Product and release resources; issue resource.
- Metrics: current SBOM per release, SBOM required fields, no exploitable vulnerabilities (respecting VEX), time to
  triage, time to fix, advisory for fixed vulnerability, security policy present, branch protection, signed commits,
  scans run on every release.

## 7. UX

Principles: the product, not the data model, is the entry point; every red item says what to do next; the user never
needs a command line.

**Onboarding wizard** (five steps, each saved, can be resumed)
1. Product: name, description, release line, support period.
2. Classification: a few questions; the result (class and conformity route) is shown with an explanation and can be
   overridden.
3. Scope: the generated CRA catalog and the controls in scope are previewed.
4. Connect: GitHub (app install), Dependency-Track (URL, key, suggested projects), issue tracker; each connection shows
   status and the first data read. Skipping is allowed; skipped sources create tasks.
5. First picture: the product page with a compliance summary and the first work queue.

**Product page** (home of a product)
- Compliance by CRA part as a clear summary with the trend, not a wall of controls.
- Sources panel with connection status and last sync.
- Releases with their snapshot and SBOM status.
- Deadlines (reporting clocks, support end).

**Work queue**: tasks and findings the user has to act on, grouped by what is due soonest, filterable by assignee. Each
entry links to the control, the evidence or the external system and offers the direct action (write the document,
add a manual statement, open in DT).

**Control view**: requirement in plain language and the CRA reference, evidence found, what is missing, manual
statements (per metric and per resource, automatic results take precedence), comments, history.

**Authoring**: templates for the documents the CRA expects (risk assessment, vulnerability disclosure policy, user
information, ...), editor, versions, review and approval, then stored as evidence.

**Vulnerability and incident handling**: shows the clocks of actively exploited vulnerabilities and severe incidents,
drafts the reports from the available data, records submission proof. Triage itself stays in DT.

**Conformity output**: technical file (Annex VII) and declaration of conformity (Annex V) generated from a snapshot,
reviewed and approved in the tool, exported as PDF and machine-readable bundle, with the content hash.

**Cross-cutting**: notifications (email, Slack, webhooks), comments, roles that match jobs (compliance manager,
engineer, auditor), accessibility and keyboard use, a guided empty state everywhere.

## 8. Evaluation and data

- Evaluation continues per control and audit scope. Product class and release feed scope selection and snapshots.
- SLA and clock evaluation is a new evaluator on top of evidence timestamps (DT and issue data), producing tasks and
  deadlines.
- Snapshot creation records referenced result ids and a hash; evidence is never copied. Retention keeps everything
  referenced by a snapshot for at least the CRA retention period.
- Performance: new tables follow the index and IDs-first query approach used for evaluation and assessment results;
  collectors submit changes only.

## 9. Operations and security

- Persistent database is the default deployment; the in-memory mode stays for demos. A Helm chart and a compose file
  with persistence are part of the product, plus backup and upgrade documentation.
- Secrets for connections are encrypted at rest with a key from the environment or a secret manager; they are write
  only through the API.
- Authorization stays per product (ToE) and audit scope; connection and secret management require the manager role.
- Observability: structured logs exist; add metrics (collector runs, sync lag, evaluation duration) and health
  endpoints for the connector runtime.
- Release process: images published per release, changelog, migration notes.

## 10. Example product

A public repository in the Confirmate organization based on the `pythonpass` demo (`cra-demo`) that is a complete,
small product:

- source, CI pipeline with tests and scans, SBOM generated in CI and uploaded to DT, releases with signed artifacts,
- `SECURITY.md`, vulnerability disclosure policy and the CRA documents,
- a few security issues in different states,
- a compose file that starts Confirmate, DT and a seeded product.

It serves as demo, documentation and end-to-end test of the connectors. Creating the public repository requires an
explicit go-ahead.

## 11. Milestones

Each milestone ends with something demonstrable and an automated test.

| M | Content | Done when |
|---|---------|-----------|
| M1 | CRA catalog v1 (Part I and II, Art. 13/14 obligations that can be evidenced), metric gaps closed or marked manual, `Product` resource, product page and work queue (without connectors) | A product can be created in the UI, the catalog scope is generated, manual statements and documents can be added, the work queue shows open items |
| M2 | GitHub connector and example product | The example product's repository settings and workflows appear as evidence and drive controls |
| M3 | DT connector, SBOM and vulnerability metrics, onboarding wizard | A DT project is linked in the wizard, SBOM and findings drive the Part II controls, webhook refresh works |
| M4 | Clocks and reporting, issue tracker, tasks and deadlines | A simulated exploited vulnerability starts the 24 h/72 h clocks and drafts the report |
| M5 | Authoring, snapshots, technical file and declaration of conformity | A release produces a verifiable technical file bundle |
| M6 | Persistent deployment, Helm chart, notifications, observability, docs | A new user installs and reaches a first picture following the quickstart only |
| M7 | CSAF through DT, signing and provenance evidence | Advisories for fixed vulnerabilities are checked automatically |

## 12. Risks

- **Catalog correctness**: wrong or incomplete mapping undermines trust. Mitigation: external review, versioned
  changelog, neutral wording.
- **LLM-based document evidence is non-deterministic.** Mitigation: manual confirmation of extracted facts before they
  count, and deterministic checks where a format exists (SBOM, CSAF).
- **DT API change** (v5 moves resources to the new API). Mitigation: connector interface, capability detection, e2e test
  against a DT container.
- **Credential handling** for connections. Mitigation: encrypted store, write-only API, minimal scopes, audit trail.
- **Scope growth.** Mitigation: milestones above, non-goals, D1 to D10.

## 13. Questions that need input from outside this document

- Who reviews the catalog content against the regulation, and when?
- Which conformity routes and product classes do the first customers need?
- Which GitHub features can the first customers use (GitHub Enterprise Server, GitLab instead)?
