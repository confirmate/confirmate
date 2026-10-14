# CRA compliance management: product design

Status: draft for discussion. Nothing in here is implemented yet.

## Goal

A company can manage the Cyber Resilience Act (CRA) compliance of a software product end to end in Confirmate:
connect the sources of the product, see what is missing, work off the open items, and hand a technical file and an
EU declaration of conformity to an authority or customer. It must be pleasant to use without knowing the CRA text or
the Confirmate data model.

We do not build an SBOM or vulnerability-matching solution. [Dependency-Track](https://dependencytrack.org) (DT) is
the source of truth for SBOMs, components and vulnerabilities. Confirmate turns its data into compliance statements,
tasks, documents and deadlines.

## Where we are

See the gap analysis in the discussion that led to this document. In short: the core (audit scopes, control states,
manual results, reports, permissions) and the ontology types for CRA exist, but nothing fills them for a real product:
no git platform, SBOM, vulnerability or issue collector, only an example catalog for Annex I Part I, no onboarding in
the UI, no persistent deployment story.

## Product model

The central object is the **product**, not the target of evaluation:

- name, version / release line, product class (default, important class I/II, critical), support period end;
- the chosen conformity route, derived from the class (self-assessment, harmonised standard, third party);
- the linked sources: repository, DT project, issue tracker, documents;
- responsible people.

A product maps to one target of evaluation and an audit scope on the generated CRA catalog, so the existing services
keep working. The product class decides which controls of the catalog are in scope.

## CRA catalog

Replace `example-data/catalogs/cra-example-catalog.json` (Annex I Part I only, 27 leaves, marked "example") by a
catalog that covers

- Annex I Part I (security properties of the product),
- Annex I Part II (vulnerability handling: SBOM, vulnerability disclosure policy, advisories, free security patches),
- Art. 13 and 14 (obligations of manufacturers, reporting of actively exploited vulnerabilities and incidents),
- Annex VII (technical documentation).

Every control gets either metrics (automatic) or is explicitly marked as manual with a description of the expected
evidence. Controls carry the CRA article reference. The mapping is reviewed against the regulation text, not only the
existing example numbering.

## Dependency-Track integration

A new collector reads from DT and submits evidence like the cloud collectors (`core/service/collection`).

Target is the current DT (v5, GA, `DependencyTrack/dependency-track` `main`; v4 stays in maintenance for some time and
is out of scope). Facts from the DT repository (verified in the source, not against a running instance yet):

- the stable REST API under `/api/v1` is still the way to get projects, BOMs, findings, VEX, policy violations and
  metrics (`ProjectResource`, `BomResource`, `FindingResource`, `VexResource`, `MetricsResource`, `AnalysisResource`);
- the new spec-first API (`api/src/main/openapi`) covers only some resources so far (e.g. `/components`,
  `/projects/{uuid}/components`, `/vuln-policies`) and has `/internal/system-capabilities`;
- CSAF (`/csaf/*`, `/advisories`) is not in a release yet.

Design:

- Read per linked project: latest BOM (CycloneDX), components, findings with their analysis state, VEX, policy
  violations, project metrics.
- Push: configure a DT webhook alert for BOM processed and new vulnerability events to refresh the compliance state
  quickly; polling is the fallback.
- The collector hides the API version behind a small interface, so the new API can be used step by step.
- CSAF is a capability that is detected at runtime. Until DT offers it, advisory metrics report "not available",
  not failure.

Ontology additions (to be proposed in `security-metrics`): richer `Vulnerability` (severity, analysis state,
published, fixed and triaged timestamps), SBOM facts (format, spec version, component count, required fields present),
and a product / release resource.

New metrics: a current SBOM exists for every release; the SBOM has the required fields; no known exploitable
vulnerabilities (respecting VEX); time to triage and to fix within the SLA; an advisory exists for fixed
vulnerabilities (when CSAF is available).

## Other sources

- Git platform (GitHub first, GitLab later): repository settings (branch protection, required reviews, signed commits,
  2FA, `SECURITY.md`, dependency update bot), workflows. Fills the existing `CodeRepository`, `Workflow` and `Job`.
- Issue tracker (GitHub issues first, then Jira): security issues, labels, open and fix times. Needs a new ontology
  type.
- Documents: the existing document analyser plus manual authoring in the UI (see below).
- Release artifacts: signing and provenance later.

## UX

- **Onboarding wizard**: create product, answer a few classification questions, get the generated scope, connect
  sources (GitHub app / OAuth, DT URL and API key with project suggestions by package URL, optional document upload).
  First compliance picture within minutes, no flags or environment variables.
- **Work queue** on the home screen: failing controls, missing documents, expiring manual statements, security
  issues beyond their SLA, upcoming deadlines; each item has an assignee, due date and status.
- **Authoring**: documents without automatic evidence (risk assessment, disclosure policy, user information) can be
  written or uploaded in the tool, versioned, reviewed and approved; manual per-metric results with evidence.
- **Vulnerability and incident handling**: track the reporting clocks of actively exploited vulnerabilities (early
  warning 24 h, notification 72 h, final report) and draft the report. DT stays the place to triage; Confirmate links
  to it.
- **Conformity output**: technical file (Annex VII) and declaration of conformity generated from the current state,
  with version and evidence links, as signed PDF and machine-readable bundle; point-in-time snapshots ("state at
  release 2.3").
- Notifications (email, Slack, webhooks) and comments on controls.

## Example product

A public repository in the Confirmate organization based on the `pythonpass` demo (`cra-demo`), with: source and
CI pipeline, an SBOM generated in CI and uploaded to DT, `SECURITY.md` and disclosure policy, CRA documents, release
artifacts and security issues. A compose file starts Confirmate and DT with it. It is the end-to-end test of
everything above and the base for demos. Creating the public repository needs a go-ahead.

## Order

1. CRA catalog, GitHub collector, example product, product dashboard with work queue.
2. Onboarding wizard, DT collector (SBOM and vulnerability flow), reporting clocks.
3. Document authoring, technical file and declaration of conformity.
4. Persistent deployment (Helm chart, database docs), notifications, snapshots, organization view, observability.
5. Issue tracker integration (GitHub, then Jira), CSAF through DT when released.

## Open questions

- Single organization first, or multi-tenant from the start? Proposal: single organization.
- Does Confirmate ever write to DT (e.g. create advisories) or stay read-only? Proposal: read-only.
- Who owns the CRA catalog content and its review against the regulation?
