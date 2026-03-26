import { confirmatize } from '$lib/api/util';
import { throwError } from './errors';
import type { AuditScope } from './orchestrator';

export interface StartEvaluationResponse {
	succesful: boolean;
}

export interface ListEvaluationResultsResponse {
	results: EvaluationResult[];
}

export type ComplianceStatus =
	| 'EVALUATION_STATUS_NOT_COMPLIANT'
	| 'EVALUATION_STATUS_COMPLIANT'
	| 'EVALUATION_STATUS_COMPLIANT_MANUALLY'
	| 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY'
	| 'EVALUATION_STATUS_PENDING'
	| 'EVALUATION_STATUS_DELEGATED';

export interface EvaluationResult {
	id: string;
	targetOfEvaluationId: string;
	status: ComplianceStatus;
	controlCatalogId: string;
	controlCategoryName?: string;
	controlId: string;
	parentControlId?: string;
	timestamp: string;
	failingAssessmentResultIds: string[];
	comment?: string;
	validUntil?: string;
}

export async function startEvaluation(auditScope: AuditScope): Promise<StartEvaluationResponse> {
	const apiUrl = confirmatize(`/v1/evaluation/evaluate/${auditScope.id}/start`);

	return fetch(apiUrl, {
		method: 'POST',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	}).then((res) => res.json());
}

export async function stopEvaluation(auditScope: AuditScope): Promise<{}> {
	const apiUrl = confirmatize(`/v1/evaluation/evaluate/${auditScope.id}/stop`);

	return fetch(apiUrl, {
		method: 'POST',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	}).then((res) => res.json());
}

export interface ListEvaluationResultsFilter {
	targetOfEvaluationId?: string;
	catalogId?: string;
	controlId?: string;
	subControls?: string;
	parentsOnly?: boolean;
	validManualOnly?: boolean;
}

export async function listEvaluationResults(
	filter?: ListEvaluationResultsFilter,
	latestByControlId = false,
	fetch = window.fetch
): Promise<EvaluationResult[]> {
	let apiUrl = confirmatize(`/v1/evaluation/results?`);

	if (filter?.targetOfEvaluationId != undefined) {
		apiUrl += `&filter.target_of_evaluation_id=${filter?.targetOfEvaluationId}`;
	}
	if (filter?.catalogId != undefined) {
		apiUrl += `&filter.catalog_id=${filter?.catalogId}`;
	}
	if (filter?.controlId != undefined) {
		apiUrl += `&filter.control_id=${filter?.controlId}`;
	}
	if (filter?.parentsOnly != undefined) {
		apiUrl += `&filter.parents_only=${filter?.parentsOnly}`;
	}
	if (latestByControlId) {
		apiUrl += `&latest_by_control_id=true`;
	}

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListEvaluationResultsResponse) => {
			return response.results;
		});
}

export async function createEvaluationResult(
	result: EvaluationResult,
	fetch = window.fetch
): Promise<EvaluationResult> {
	const apiUrl = confirmatize(`/v1/evaluation/results`);

	return fetch(apiUrl, {
		method: 'POST',
		body: JSON.stringify(result),
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json());
}
