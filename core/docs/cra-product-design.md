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
| D3 | Confirmate does not maintain SBOM or vulnerability tooling. It reads SBOMs, vulnerabilities and VEX from where they live: Dependency-Track (DT) as the first system, or directly from standard files. Confirmate never writes to these systems. | We do not maintain an SBOM solution; read-only keeps the trust boundary simple. |
| D4 | Target DT v5 (`DependencyTrack/dependency-track` `main`). v4 is out of scope. | v5 is GA and the future; v4 is in maintenance. |
| D5 | CSAF in DT is an optional capability, detected at runtime. | It is not in a DT release yet. |
| D6 | Connections to external systems are configured through the Confirmate API/UI and run server-side. No flags/env vars for users. | Required for a guided onboarding; credentials are stored once, encrypted. |
| D7 | The CRA catalog is maintained in the `security-metrics` repository next to the ontology and metrics, versioned, and reviewed against the regulation. | One place for compliance content; the repo already ships catalogs. |
| D8 | Automatic evidence wherever possible; everything else is an explicit manual control with a defined evidence type (document, statement). No control without either. | Prevents "silent gaps" in the catalog. |
| D9 | Documents and statements without automatic evidence are authored or uploaded in Confirmate and stored as evidence with versions and approval. | The user manages compliance in the tool, not next to it. |
| D10 | Evidence history is append-only and snapshots are first class (state at a release or a date). | Audits ask "what was true when you shipped 2.3". |
| D11 | Integrate through open formats first (CycloneDX, SPDX, VEX, CSAF, SARIF, OSV, in-toto/SLSA attestations, OpenSSF Scorecard results). A system-specific connector is built only where we need state that has no file format (git platform settings, issue timelines, DT project state). | One generic ingest serves many tools; connectors stay few and maintainable. |
| D13 | The git platform integration is platform-neutral with GitLab (including self-hosted) and GitHub as first providers. The example product is a real public project that is available on both. Reports to authorities are prepared and proven in Confirmate (draft, submission proof upload); there is no integration with an authority platform. | Many customers run their own GitLab; GitHub is common for open source and suppliers. An authority API is not needed to manage compliance. |
| D14 | We do not ship a git platform in the demo. The cloud versions (gitlab.com, github.com) stand in for self-managed instances: they serve the live profiles and are the source of recordings for a replay server. A real GitLab CE container is an optional heavy profile for self-managed edge cases. | GitLab is too heavy to ship by default; the cloud and self-managed APIs are the same; a different platform (Forgejo) would not test the GitLab connector. |
| D15 | There are two example products: one hosted on GitHub, one on GitLab. | They show the multi-product view, different classes and stacks, and give each connector a real project. |
| D12 | The demo stack is self-contained and shows every integration that the product claims, with seeded data and a scripted scenario. | A capability that is not in the demo does not exist for a prospect. |

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

### Integration landscape

Principle (D11): formats before tools. Priority: **P1** needed for a credible first product, **P2** next, **P3**
later or on demand. "Auto" fills controls; "trigger" starts a workflow (for example a reporting clock).

| Area | Integration point | Format / interface | Serves (CRA) | Prio | Kind |
|------|-------------------|--------------------|--------------|------|------|
| SBOM | Dependency-Track (v5) | REST `/api/v1`, CycloneDX | Part II (1), Art. 13 due diligence | P1 | connector |
| SBOM | SBOM files from CI or an OCI registry (generated by e.g. Syft, cdxgen, Trivy, language plugins) | CycloneDX, SPDX | Part II (1) | P1 | generic ingest |
| Vulnerabilities | VEX statements | CycloneDX VEX, OpenVEX, CSAF VEX | Part I (2)(a), Part II (2),(4) | P1 | generic ingest |
| Vulnerabilities | Scanner results (Trivy, Grype, OSV-Scanner) | OSV-style JSON, CycloneDX | Part I (2)(a) | P2 | generic ingest |
| Vulnerabilities | Exploitation signals: CISA KEV, EPSS; EU vulnerability database | public feeds | Art. 14 trigger "actively exploited" | P1 | trigger |
| Vulnerabilities | Dependabot / GitHub security alerts; GitLab vulnerability report and dependency list (Ultimate tier only) | platform APIs | Part I (2)(a), Part II (2) | P2 | via git connector |
| Advisories | CSAF documents of the product's own advisories; DT advisories later | CSAF 2.x; validator | Part II (4),(5),(6),(8) | P1 | generic ingest (CSAF discovery exists) |
| Source | GitLab (gitlab.com and self-managed) | GitLab REST API v4 | Part I (2)(b),(c); Part II (3),(5)-(7) | P1 | connector |
| Source | GitHub (App) | GitHub REST/GraphQL | same | P1 | connector |
| Source | Forgejo/Gitea, Bitbucket | respective APIs | same | P3 | connector |
| Source | OpenSSF Scorecard (branch protection, code review, signed releases, SAST, pinned dependencies, security policy, token permissions, dependency update tool, fuzzing, ...) | Scorecard JSON; public API for public repos | Part I (2)(a),(c),(j); Part II (3),(5),(7) | P1 | generic ingest |
| Code security | SAST and secret scanning results (CodeQL, Semgrep, Sonar, GitHub code scanning) | SARIF | Part I (2) general, Part II (3) | P1 | generic ingest |
| Code security | In-house static analysis (CPG/Codyze), crypto inventory (CBOM from cbomkit) | existing collector, CycloneDX CBOM | Part I (2)(b),(d),(e),(f),(g),(l),(m) | P1 | existing + generic ingest |
| Supply chain | Signatures and provenance (cosign/sigstore, SLSA provenance, GitHub artifact attestations) | in-toto attestations | Part II (7), Art. 13 | P2 | generic ingest |
| CI/CD | GitLab CI / GitHub Actions run history, required checks, test and coverage reports | CI API, JUnit/coverage files | Part II (3) | P1 | via git connector |
| Issues | GitLab and GitHub issues, GitHub Security Advisories and private vulnerability reporting, GitLab confidential issues | platform APIs | Part II (2),(5),(6), Art. 14 | P1 | via git connector |
| Issues | Jira, Linear | respective APIs | Part II (2) | P2 | connector |
| Documents | Documents in the repository (docs-as-code), uploaded files, Confluence/SharePoint | Markdown/PDF, APIs | Annex VII, Art. 13 | P1 repo+upload / P3 wikis | analyser + authoring |
| Infrastructure | Cloud and Kubernetes the product runs on | existing cloud collectors | Part I (2)(d),(e),(h),(l) | exists | existing |
| Reporting | Authority reporting platform | no integration: report drafts are prepared in Confirmate, the user submits them and uploads the proof | Art. 14 | P1 | output |
| Notification | Email, Slack/Teams, outgoing webhooks | SMTP, webhooks | tasks, clocks | P1 | output |
| Identity | OIDC provider (Keycloak, Entra, Okta) | OIDC | access | P2 | existing basis |

Consequences:

- A **generic evidence ingest** ("bring your own result") comes before most connectors: an endpoint and a CI snippet
  that accepts CycloneDX, SPDX, VEX, SARIF, CSAF, Scorecard JSON and attestations for a product and release, validates
  them and maps them into the ontology. Almost every row marked "generic ingest" is then done by a CI step and no
  connector at all. It also covers customers who do not run DT.
- **Three connectors** carry the state that has no file format: git platform (repository settings, pipelines, issues,
  security advisories; GitLab and GitHub), Dependency-Track, and issue tracker (Jira).
- In-house tools stay relevant as evidence sources: CPG/Codyze for code facts, cbomkit for crypto inventory (CBOM),
  the CSAF validator and tooling for advisories.
- The Art. 14 trigger does not need a customer to tell us that something is exploited: KEV/EPSS matches against the
  product's components start the clocks as a suggestion that a person confirms.

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

### Git platform (GitLab and GitHub)

One platform-neutral model, two providers behind one interface (`GitPlatform`): repository settings (protected
branches, required reviews and approvals, signed commits or push rules, 2FA requirement, `SECURITY.md`, dependency
update bot), pipelines/workflows and their recent runs, releases and their assets, security issues and advisories.
Fills `CodeRepository`, `Workflow`, `Job`. Authentication: GitHub App installation; GitLab project or group access
token with read scopes.

Self-managed GitLab is the main case. The provider is part of the product; adapting it to a specific customer
instance (version, certificates, tiers, network, access policy) is a service engagement, and most of the work is
already done by then. The provider has to cope with:

- **Editions and tiers:** several settings (approval rules, push rules, vulnerability report, dependency list) exist
  only in paid tiers or in some versions. A missing feature is reported as "not available in this edition", which
  leads to a manual control or a CI-based alternative, never to a failed connection. The generic ingest of SBOM, SARIF
  and Scorecard results from CI artifacts covers the free tier.
- **Network and trust:** the connector runs where it can reach the instance (customer side, or a runner); custom CA
  certificates and an instance URL are part of the connection; no inbound access to the instance is needed.
- **Least privilege:** read-only scopes only, shown to the user before the token is created.
- **Versions:** the supported GitLab versions are stated and tested against the recorded API.

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

## 10. Example product and demo stack

### Example products (D15)

Two small, complete products with the same pipeline templates but different properties, so that the organization view
and the classification have something to show:

| | Product A | Product B |
|---|-----------|-----------|
| Based on | `pythonpass` (`cra-demo`), a password manager | a second small product, for example a web service in Go or Node with a container image |
| Hosted on | GitHub (Confirmate organization) | gitlab.com (Confirmate group) |
| CI | GitHub Actions | GitLab CI |
| CRA class | likely important class I (password managers are listed in Annex III; to be verified), so a stricter conformity route is shown | default class |
| Ecosystem | PyPI | different one (Go modules or npm), plus container image and its SBOM |
| Compliance picture | mostly compliant with a few visible gaps | deliberately problematic: stale dependency, no signed releases, missing policy |

Each product contains: source, a CI pipeline with tests, SAST and an SBOM, releases with signed artifacts and
provenance, `SECURITY.md` and a vulnerability disclosure policy, the CRA documents, and security issues and
advisories in different states. Each has the pipeline definition of its own platform in the repository; nothing is
mirrored between the platforms.

Creating the public projects, and applying for GitLab's open-source program (which may give the project a paid tier for
the demo; conditions to be checked), require an explicit go-ahead.

### Demo stack (D12)

`docker compose --profile cra up` starts everything below. Data is seeded so that the first screen already tells a
story; no internet access and no token are needed after the images are pulled. Further profiles `cra-gitlab` and
`cra-github` run the connectors live against the cloud-hosted example products with a token, for demos that should
show real changes. A heavy profile `cra-gitlab-selfmanaged` starts a real GitLab CE container with a runner and a
seed container, for self-managed edge cases (older version, custom CA, features missing in the edition); it takes
minutes to start and needs several GB of memory, so it is not part of the default. We do not ship a git platform by
default (D14). Forgejo or Gitea would only be added together with a connector for them; they cannot stand in for
GitLab, because their API differs.

Testing: end-to-end tests run against the recorded replay data, which is fast and deterministic. A scheduled contract
test against gitlab.com and github.com detects API drift and triggers a re-recording; the self-managed profile runs
per release.

| Component | Role in the demo | Integration shown |
|-----------|------------------|-------------------|
| Confirmate core, UI, connector runtime | the product, with a persistent PostgreSQL | everything |
| Dependency-Track v5 (+ its PostgreSQL) | holds the SBOM of the example product | DT connector, VEX, webhook refresh |
| Git platform replay servers (GitLab and GitHub) | serve responses recorded from the two example products (settings, pipelines, issues, advisories, alerts), offline and without a token | GitLab and GitHub connectors, organization view with two products |
| CI job (seed container) | runs Syft/cdxgen, Trivy or OSV-Scanner, Semgrep, Scorecard against the example repo and uploads results | generic ingest: CycloneDX, SARIF, OSV, Scorecard |
| CPG/Codyze collector, document analyser | code and document facts | existing collectors |
| CSAF provider and validator | serves the product's advisories (from the `cra-demo` CSAF generator) | CSAF discovery and validation |
| Exploitation feed mock | serves a KEV-style feed that lists a component of the example product | Art. 14 clock suggestion |
| Mailpit and a webhook sink | receive notifications | notifications |
| Keycloak (optional profile) | external OIDC login | identity |
| Prometheus/Grafana (optional profile) | operations view | observability |

### Scripted scenario

1. A fresh start shows the organization with both example products, each with a different compliance picture, and a
   work queue. Product A shows the stricter conformity route of its class.
2. A new vulnerability appears in a component: DT updates, Confirmate shows the failing control and a task; the KEV
   mock lists it, the 24 h clock is suggested, a person confirms, the report draft is created and the user uploads
   the submission proof.
3. The fix is released: new release, SBOM and provenance arrive, the control turns green, an advisory is published and
   validated.
4. The user adds a manual statement and a document for a control without automatic evidence.
5. A snapshot is taken for the release and the technical file and declaration of conformity are generated.

The scenario is a script that is also run in CI as an end-to-end test of the integrations.

## 11. Milestones

Each milestone ends with something demonstrable and an automated test.

| M | Content | Done when |
|---|---------|-----------|
| M1 | CRA catalog v1 (Part I and II, Art. 13/14 obligations that can be evidenced), metric gaps closed or marked manual, `Product` resource, product page and work queue (without connectors) | A product can be created in the UI, the catalog scope is generated, manual statements and documents can be added, the work queue shows open items |
| M2 | Generic evidence ingest (CycloneDX, SPDX, VEX, SARIF, CSAF, Scorecard), platform-neutral git connector with GitLab and GitHub providers, both example products, first demo profile with replay data | The example product's repository settings, workflows, SBOM and scan results appear as evidence and drive controls |
| M3 | DT connector, SBOM and vulnerability metrics, KEV/EPSS signals, onboarding wizard, demo stack with DT | A DT project is linked in the wizard, SBOM and findings drive the Part II controls, webhook refresh works, scenario steps 1 to 2 run |
| M4 | Clocks and report drafts with proof upload, issue tracker (Jira), tasks and deadlines, notifications | A simulated exploited vulnerability starts the 24 h/72 h clocks, drafts the report and notifies the assignee |
| M5 | Authoring, snapshots, technical file and declaration of conformity | A release produces a verifiable technical file bundle |
| M6 | Persistent deployment, Helm chart, notifications, observability, docs | A new user installs and reaches a first picture following the quickstart only |
| M7 | CSAF through DT, signing and provenance evidence, self-managed GitLab profile and contract tests, Forgejo/Gitea provider on demand, Keycloak profile | Advisories for fixed vulnerabilities are checked automatically; the full scenario runs in CI |

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
- Which GitLab versions and tiers do the first customers run, and do they allow a connector with a read token?
- Which issue tracker do the first customers use (GitLab issues, Jira, other)?
