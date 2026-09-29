import createClient from 'openapi-fetch';
import type { paths as OrchestratorPaths } from './openapi/orchestrator';
import type { paths as AssessmentPaths } from './openapi/assessment';
import type { paths as EvidencePaths } from './openapi/evidence';
import type { paths as EvaluationPaths } from './openapi/evaluation';

/** The `Authorization: Bearer <token>` header for the current user, if logged in. Exported so
 * callers that need a raw `fetch` outside the generated openapi clients (e.g. file downloads,
 * which can't go through a client that always parses the response as JSON) can still authenticate. */
export function authHeaders(): HeadersInit {
	if (typeof globalThis.localStorage === 'undefined') return {};
	const token = globalThis.localStorage.getItem('token');
	return token ? { Authorization: `Bearer ${token}` } : {};
}

export function orchestratorClient(fetch?: typeof globalThis.fetch) {
	return createClient<OrchestratorPaths>({ headers: authHeaders(), fetch });
}

export function assessmentClient(fetch?: typeof globalThis.fetch) {
	return createClient<AssessmentPaths>({ headers: authHeaders(), fetch });
}

export function evidenceClient(fetch?: typeof globalThis.fetch) {
	return createClient<EvidencePaths>({ headers: authHeaders(), fetch });
}

export function evaluationClient(fetch?: typeof globalThis.fetch) {
	return createClient<EvaluationPaths>({ headers: authHeaders(), fetch });
}
