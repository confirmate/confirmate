import { orchestratorClient } from '$lib/api/client';
import type {
	SchemaAssessmentResult,
	SchemaAuditScope,
	SchemaCatalog,
	SchemaControl,
	SchemaControlInScope,
	SchemaEvaluationResult,
	SchemaRiskAssessment,
	SchemaTargetOfEvaluation,
	SchemaUser
} from '$lib/api/openapi/orchestrator';
import { bucketControlsByCategory, flattenControls } from '$lib/auditScope';
import { computeScope } from './scope';
import type {
	AcceptedRisk,
	Check,
	GuidedSource,
	NewStatement,
	NewTarget,
	PartProgress,
	ScopeChange,
	Progress,
	RequirementDetail,
	RequirementStatus,
	RequirementSummary,
	Task,
	TargetDetail,
	TargetSummary
} from './source';

/** Everything that is known about the checking of one target of evaluation against one framework. */
interface Snapshot {
	toe: SchemaTargetOfEvaluation;
	scope?: SchemaAuditScope;
	catalog?: SchemaCatalog;
	assessment?: SchemaRiskAssessment;
	parts: (PartProgress & { items: RequirementSummary[] })[];
	progress: Progress;
	assignees: Map<string, string>;
}

const typeLabels: Record<string, string> = {
	TARGET_TYPE_PRODUCT: 'Product',
	TARGET_TYPE_CLOUD: 'Cloud service',
	TARGET_TYPE_ORGANIZATION: 'Organisation'
};

const classLabels: Record<string, string> = {
	CLASSIFICATION_DEFAULT: 'default class',
	CLASSIFICATION_IMPORTANT_I: 'important class I',
	CLASSIFICATION_IMPORTANT_II: 'important class II',
	CLASSIFICATION_CRITICAL: 'critical'
};

const productTypes: Record<string, string> = {
	'Software application': 'PRODUCT_TYPE_SOFTWARE_APPLICATION',
	'Library or component': 'PRODUCT_TYPE_LIBRARY',
	'Firmware or embedded software': 'PRODUCT_TYPE_FIRMWARE',
	'Online service with a client': 'PRODUCT_TYPE_ONLINE_SERVICE_WITH_CLIENT'
};

/**
 * The generated types of the API mark fields as required that the server sets itself (IDs, timestamps, versions). This
 * lets a request body leave them out; the server validates the request anyway.
 */
function request(value: object): never {
	return value as never;
}

/** Turns an error response of the API into an error with a readable message. */
function fail(what: string, error: unknown): never {
	const message = (error as { message?: string } | undefined)?.message;
	throw new Error(message ? `${what}: ${message}` : what);
}

function statusOf(result?: SchemaEvaluationResult): RequirementStatus {
	switch (result?.status) {
		case 'EVALUATION_STATUS_COMPLIANT':
		case 'EVALUATION_STATUS_COMPLIANT_MANUALLY':
			return 'met';
		case 'EVALUATION_STATUS_NOT_COMPLIANT':
		case 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY':
			return 'attention';
		default:
			return 'open';
	}
}

function sum(parts: Progress[]): Progress {
	return parts.reduce(
		(a, p) => ({ met: a.met + p.met, attention: a.attention + p.attention, open: a.open + p.open, total: a.total + p.total }),
		{ met: 0, attention: 0, open: 0, total: 0 }
	);
}

/** The first audit scope of the target of evaluation, preferring the one of the given catalog. */
function pickScope(scopes: SchemaAuditScope[], catalogId?: string): SchemaAuditScope | undefined {
	return scopes.find((s) => s.catalogId === catalogId) ?? scopes[0];
}

async function loadSnapshot(toe: SchemaTargetOfEvaluation, scopes: SchemaAuditScope[], users: SchemaUser[]): Promise<Snapshot> {
	const client = orchestratorClient();
	const assignees = new Map(
		users.map((u) => [u.id ?? '', [u.firstName, u.lastName].filter(Boolean).join(' ') || u.username || u.id || ''])
	);
	const scope = pickScope(scopes.filter((s) => s.targetOfEvaluationId === toe.id));

	if (!scope) {
		return { toe, parts: [], progress: { met: 0, attention: 0, open: 0, total: 0 }, assignees };
	}

	const [catalogRes, controlsRes, inScopeRes, resultsRes, assessmentsRes] = await Promise.all([
		client.GET('/v1/orchestrator/catalogs/{catalogId}', { params: { path: { catalogId: scope.catalogId } } }),
		client.GET('/v1/orchestrator/controls', { params: { query: { 'filter.catalogId': scope.catalogId, pageSize: 1500 } } }),
		client.GET('/v1/orchestrator/controls_in_scope', { params: { query: { 'filter.auditScopeId': scope.id, pageSize: 1500 } } }),
		client.GET('/v1/orchestrator/evaluation_results', {
			params: {
				query: {
					'filter.targetOfEvaluationId': toe.id,
					'filter.auditScopeId': scope.id,
					latestByControlId: true,
					pageSize: 1500
				}
			}
		}),
		client.GET('/v1/orchestrator/risk_assessments', {
			params: { query: { 'filter.auditScopeId': scope.id, pageSize: 50 } }
		})
	]);

	const catalog = catalogRes.data;
	const topLevel: SchemaControl[] = controlsRes.data?.controls ?? [];
	const inScope = new Map<string, SchemaControlInScope>(
		(inScopeRes.data?.controlsInScope ?? []).map((c) => [c.controlId ?? '', c])
	);
	const results = new Map<string, SchemaEvaluationResult>(
		(resultsRes.data?.results ?? []).map((r) => [r.controlId ?? '', r])
	);
	const assessment = (assessmentsRes.data?.riskAssessments ?? []).find((a) => a.status === 'RISK_ASSESSMENT_STATUS_APPROVED');

	const byCategory = bucketControlsByCategory(catalog?.categories ?? [], topLevel);
	const parts = Object.entries(byCategory)
		.map(([name, controls]) => {
			const items: RequirementSummary[] = flattenControls(controls)
				.filter((c) => (c.controls ?? []).length === 0 && inScope.has(c.id ?? ''))
				.map((c) => ({
					id: c.id ?? '',
					partId: name,
					title: c.name ?? c.shortName ?? '',
					reference: c.shortName ?? '',
					status: statusOf(results.get(c.id ?? ''))
				}));
			const count = (s: RequirementStatus) => items.filter((i) => i.status === s).length;
			return { id: name, label: name, met: count('met'), attention: count('attention'), open: count('open'), total: items.length, items };
		})
		.filter((p) => p.total > 0);

	return { toe, scope, catalog, assessment, parts, progress: sum(parts), assignees };
}

function frameworkLabel(snapshot: Snapshot): string {
	if (!snapshot.catalog) return 'No framework yet';
	const name = snapshot.catalog.shortName ?? snapshot.catalog.name ?? '';
	const cls = snapshot.toe.product?.classification;
	return cls && classLabels[cls] ? `${name}, ${classLabels[cls]}` : name;
}

function summary(snapshot: Snapshot): TargetSummary {
	const { toe } = snapshot;
	const release = toe.product?.currentRelease;
	const review = snapshot.assessment?.reviewDueAt;

	return {
		id: toe.id ?? '',
		name: toe.name ?? '',
		subtitle: [toe.description, release ? `release ${release}` : ''].filter(Boolean).join(', '),
		typeLabel: typeLabels[toe.targetType ?? ''] ?? 'Target of Evaluation',
		frameworkLabel: frameworkLabel(snapshot),
		tags: [],
		progress: snapshot.progress,
		nextDeadline: review ? `Review of the risk assessment, due ${new Date(review).toLocaleDateString()}` : undefined
	};
}

function tasksOf(snapshot: Snapshot): Task[] {
	const inScopeIds = new Map<string, string | undefined>();
	void inScopeIds;

	return snapshot.parts
		.flatMap((p) => p.items.map((i) => ({ ...i, part: p.label })))
		.filter((i) => i.status !== 'met')
		.sort((a, b) => (a.status === b.status ? 0 : a.status === 'attention' ? -1 : 1))
		.map((i) => ({
			id: i.id,
			title: i.title,
			detail: `${i.part}. ${i.status === 'attention' ? 'A check is not met.' : 'There is no result yet.'}`,
			urgency: i.status === 'attention' ? ('attention' as const) : ('open' as const),
			assignee: 'Unassigned',
			requirementId: i.id
		}));
}

async function loadAll() {
	const client = orchestratorClient();
	const [toesRes, scopesRes, usersRes] = await Promise.all([
		client.GET('/v1/orchestrator/targets_of_evaluation', {}),
		client.GET('/v1/orchestrator/audit_scopes', { params: { query: { pageSize: 500 } } }),
		client.GET('/v1/users', {})
	]);
	if (toesRes.error) fail('Could not load the Targets of Evaluation', toesRes.error);

	return {
		toes: toesRes.data?.targetsOfEvaluation ?? [],
		scopes: scopesRes.data?.auditScopes ?? [],
		users: usersRes.data?.users ?? []
	};
}

function checkOf(metric: NonNullable<SchemaControl['metrics']>[number], results: SchemaAssessmentResult[]): Check {
	const own = results.filter((r) => r.metricId === metric.id);
	const failing = own.find((r) => !r.compliant);
	const manual = own.some((r) => r.manual) || !metric.implementation;
	const status: Check['status'] = own.length === 0 ? 'open' : failing ? 'not-met' : 'met';
	const shown = failing ?? own[0];

	let text = metric.description ?? metric.name ?? '';
	if (shown?.complianceComment) text = shown.complianceComment;
	else if (own.length === 0) text = manual ? 'Nobody has answered this yet. Answer it with a short statement.' : 'There is no result from the source yet.';

	return {
		id: metric.id ?? '',
		title: metric.name ?? metric.id ?? '',
		kind: manual ? 'statement' : 'automatic',
		status,
		text,
		note: shown?.validUntil ? `Valid until ${new Date(shown.validUntil).toLocaleDateString()}.` : undefined,
		metricId: manual ? metric.id : undefined,
		actions: manual ? [{ id: 'statement', label: own.length === 0 ? 'Add a statement' : 'Update the statement', primary: own.length === 0 }] : []
	};
}

export const realSource: GuidedSource = {
	async listTargets() {
		const { toes, scopes, users } = await loadAll();
		const snapshots = await Promise.all(toes.map((toe) => loadSnapshot(toe, scopes, users)));
		return snapshots.map(summary);
	},

	async getTarget(id) {
		const { toes, scopes, users } = await loadAll();
		const toe = toes.find((t) => t.id === id);
		if (!toe) return undefined;

		const snapshot = await loadSnapshot(toe, scopes, users);
		const product = toe.product;

		return {
			...summary(snapshot),
			chips: [
				product?.currentRelease ? `Release ${product.currentRelease}` : '',
				product?.supportUntil ? `Supported until ${new Date(product.supportUntil).toLocaleDateString()}` : ''
			].filter(Boolean),
			tasks: tasksOf(snapshot),
			parts: snapshot.parts.map(({ items: _items, ...p }) => p),
			sources: [{ name: 'Sources', state: 'none', detail: 'Not connected yet' }],
			requirements: snapshot.parts.flatMap((p) => p.items)
		} satisfies TargetDetail;
	},

	async getRequirement(targetId, requirementId) {
		const client = orchestratorClient();
		const target = await realSource.getTarget(targetId);
		const summaryItem = target?.requirements.find((r) => r.id === requirementId);
		if (!target || !summaryItem) return undefined;

		const { data: control } = await client.GET('/v1/orchestrator/controls/{controlId}', {
			params: { path: { controlId: requirementId } }
		});
		const metrics = control?.metrics ?? [];
		const resultsRes = metrics.length
			? await client.GET('/v1/orchestrator/assessment_results', {
					params: {
						query: {
							'filter.targetOfEvaluationId': targetId,
							'filter.metricIds': metrics.map((m) => m.id ?? ''),
							latestByResourceId: true,
							pageSize: 1000
						}
					}
				})
			: undefined;

		return {
			...summaryItem,
			plain: control?.description ?? '',
			scopeReason: 'In scope',
			checks: metrics.map((m) => checkOf(m, resultsRes?.data?.results ?? [])),
			history: [],
			comments: []
		} satisfies RequirementDetail;
	},

	async createTarget(input: NewTarget) {
		const client = orchestratorClient();
		const { profile, details, choiceId, answers } = input;

		if (!details.name?.trim()) throw new Error('Please give the Target of Evaluation a name.');

		// 1. The Target of Evaluation with its product details. Only products have a guided setup so far.
		const choice = profile.choices.find((c) => c.id === choiceId);
		const toeRes = await client.POST('/v1/orchestrator/targets_of_evaluation', {
			body: request({
				name: details.name.trim(),
				description: details.description,
				targetType: profile.targetType,
				configuredMetrics: [],
				product: {
					classification: (choice?.attributes?.classification ?? 'CLASSIFICATION_UNSPECIFIED') as 'CLASSIFICATION_DEFAULT',
					productType: (productTypes[details.kind] ?? 'PRODUCT_TYPE_UNSPECIFIED') as 'PRODUCT_TYPE_SOFTWARE_APPLICATION',
					currentRelease: details.version,
					supportUntil: details.support ? new Date(`${details.support}T00:00:00Z`).toISOString() : undefined
				}
			})
		});
		const toe = toeRes.data;
		if (!toe?.id) fail('Could not create the Target of Evaluation', toeRes.error);

		// 2. The audit scope on the catalog of the framework. It brings all controls of the catalog into scope.
		const scopeRes = await client.POST('/v1/orchestrator/audit_scopes', {
			body: request({
				name: profile.name,
				targetOfEvaluationId: toe.id,
				catalogId: profile.catalogId,
				status: 'AUDIT_SCOPE_STATUS_SETUP'
			})
		});
		const scope = scopeRes.data;
		if (!scope?.id) fail('Could not create the audit scope', scopeRes.error);

		// 3. The risk assessment with the scope decisions of the questionnaire, approved right away.
		const controlsRes = await client.GET('/v1/orchestrator/controls', {
			params: { query: { 'filter.catalogId': profile.catalogId, pageSize: 1500 } }
		});
		const byShortName = new Map((controlsRes.data?.controls ?? []).map((c) => [c.shortName ?? '', c.id ?? '']));
		const groups = computeScope(profile.questionnaire, answers, input.decisions);
		const decisions: { controlId: string; inScope: boolean; justification: string; source: string }[] = [];
		for (const g of groups) {
			for (const row of g.rows) {
				const controlId = byShortName.get(row.requirement.id);
				if (!controlId || row.treatment === 'applies') continue;
				// An accepted residual risk keeps the requirement in scope; the backend has no field for it yet, so the reason
				// carries it.
				decisions.push({
					controlId,
					inScope: row.treatment === 'risk',
					justification: row.treatment === 'risk' ? `Accepted residual risk: ${row.reason}` : row.reason,
					source: row.source === 'rule' ? 'SCOPE_DECISION_SOURCE_RULE' : 'SCOPE_DECISION_SOURCE_MANUAL'
				});
			}
		}

		const assessmentRes = await client.POST('/v1/orchestrator/risk_assessments', {
			body: request({
				targetOfEvaluationId: toe.id,
				auditScopeId: scope.id,
				method: 'RISK_ASSESSMENT_METHOD_BUILT_IN',
				summary: 'Quick check in the setup wizard.',
				answers,
				decisions
			})
		});
		const assessment = assessmentRes.data;
		if (!assessment?.id) fail('Could not store the risk assessment', assessmentRes.error);

		const approveRes = await client.POST('/v1/orchestrator/risk_assessments/{riskAssessmentId}/approve', {
			params: { path: { riskAssessmentId: assessment.id } },
			body: { riskAssessmentId: assessment.id }
		});
		if (approveRes.error) fail('Could not approve the risk assessment', approveRes.error);

		return { id: toe.id };
	},

	async addStatement(statement: NewStatement) {
		const { error } = await orchestratorClient().POST('/v1/orchestrator/manual_assessment_results', {
			body: request({
				targetOfEvaluationId: statement.targetId,
				metricId: statement.metricId,
				compliant: statement.compliant,
				comment: statement.comment,
				validUntil: statement.validUntil ? new Date(`${statement.validUntil}T23:59:59`).toISOString() : undefined
			})
		});
		if (error) fail('Could not save the statement', error);
	},

	async changeScope(change: ScopeChange) {
		const client = orchestratorClient();
		const { scopes } = await loadAll();
		const scope = scopes.find((x) => x.targetOfEvaluationId === change.targetId);
		if (!scope) throw new Error('This Target of Evaluation has no audit scope yet.');

		const res = await client.GET('/v1/orchestrator/risk_assessments', {
			params: { query: { 'filter.auditScopeId': scope.id, pageSize: 50 } }
		});
		const previous = (res.data?.riskAssessments ?? []).find((a) => a.status === 'RISK_ASSESSMENT_STATUS_APPROVED');

		// The new version keeps every earlier decision except the one that is replaced.
		const decisions = (previous?.decisions ?? []).filter((d) => d.controlId !== change.requirementId);
		decisions.push({
			controlId: change.requirementId,
			inScope: change.treatment === 'risk',
			justification: change.treatment === 'risk' ? `Accepted residual risk: ${change.reason}` : change.reason,
			source: 'SCOPE_DECISION_SOURCE_MANUAL'
		});

		const created = await client.POST('/v1/orchestrator/risk_assessments', {
			body: request({
				targetOfEvaluationId: change.targetId,
				auditScopeId: scope.id,
				method: previous?.method ?? 'RISK_ASSESSMENT_METHOD_BUILT_IN',
				summary: 'Changed in the requirement view.',
				answers: previous?.answers ?? {},
				decisions
			})
		});
		if (!created.data?.id) fail('Could not store the risk assessment', created.error);

		const approved = await client.POST('/v1/orchestrator/risk_assessments/{riskAssessmentId}/approve', {
			params: { path: { riskAssessmentId: created.data.id } },
			body: { riskAssessmentId: created.data.id }
		});
		if (approved.error) fail('Could not approve the risk assessment', approved.error);
	},

	async listAcceptedRisks(targetId) {
		const client = orchestratorClient();
		const { scopes, users } = await loadAll();
		const scope = scopes.find((x) => x.targetOfEvaluationId === targetId);
		if (!scope) return [];

		const [assessments, controls] = await Promise.all([
			client.GET('/v1/orchestrator/risk_assessments', { params: { query: { 'filter.auditScopeId': scope.id, pageSize: 50 } } }),
			client.GET('/v1/orchestrator/controls', { params: { query: { 'filter.catalogId': scope.catalogId, pageSize: 1500 } } })
		]);
		const approved = (assessments.data?.riskAssessments ?? []).find((a) => a.status === 'RISK_ASSESSMENT_STATUS_APPROVED');
		if (!approved) return [];

		const byId = new Map(flattenControls(controls.data?.controls ?? []).map((c) => [c.id ?? '', c]));
		const by = users.find((u) => u.id === approved.approvedBy);
		const name = by ? [by.firstName, by.lastName].filter(Boolean).join(' ') || by.username || '' : 'Unknown';
		const prefix = 'Accepted residual risk: ';

		return (approved.decisions ?? [])
			.filter((d) => d.inScope && d.justification?.startsWith(prefix))
			.map(
				(d): AcceptedRisk => ({
					id: d.controlId ?? '',
					requirement: byId.get(d.controlId ?? '')?.name ?? d.controlId ?? '',
					reference: byId.get(d.controlId ?? '')?.shortName ?? '',
					text: (d.justification ?? '').slice(prefix.length),
					acceptedBy: name,
					acceptedAt: approved.approvedAt ? new Date(approved.approvedAt).toLocaleDateString() : '',
					reviewBy: approved.reviewDueAt ? new Date(approved.reviewDueAt).toLocaleDateString() : ''
				})
			);
	}
};
