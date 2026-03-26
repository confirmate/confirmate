import type {
	AssessmentResult,
	Metric,
	MetricConfiguration,
	MetricImplementation
} from './assessment';
import { throwError } from '$lib/api/errors';
import { confirmatize } from '$lib/api/util';

// generated types from OpenAPI spec
import type {
	SchemaRuntime,
	SchemaDependency,
	SchemaTargetOfEvaluation,
	SchemaAuditScope,
	SchemaCatalog,
	SchemaCategory,
	SchemaControl,
	SchemaCertificate,
	SchemaState,
	SchemaListCatalogsResponse,
	SchemaListControlsResponse,
	SchemaListCertificatesResponse,
	SchemaListAssessmentResultsResponse,
	SchemaListTargetsOfEvaluationResponse,
	SchemaListAuditScopesResponse,
	SchemaListMetricConfigurationResponse,
	SchemaListMetricsResponse
} from '$lib/api/openapi/orchestrator';

// keep Tag and ControlInScope manually as they are not part of the OpenAPI spec
export interface Tag {
	tag: object;
}

export interface ControlInScope {
	auditScopeTargetOfEvaluationId: string;
	auditScopeCatalogId: string;
	controlId: string;
	controlCategoryName: string;
	controlCategoryCatalogId: string;
	monitoringStatus: string;
}

// alias definitions for generated schemas
export type Runtime = SchemaRuntime;
export type Dependency = SchemaDependency;
export type TargetOfEvaluation = SchemaTargetOfEvaluation;
export type AuditScope = SchemaAuditScope;
export type Catalog = SchemaCatalog;
export type Category = SchemaCategory;
export type Control = SchemaControl;
export type Certificate = SchemaCertificate;
export type State = SchemaState;

export type ListCatalogsResponse = SchemaListCatalogsResponse;
export type ListControlsResponse = SchemaListControlsResponse;
export type ListCertificatesResponse = SchemaListCertificatesResponse;
export type ListAssessmentResultsResponse = SchemaListAssessmentResultsResponse;
export type ListTargetsOfEvaluationResponse = SchemaListTargetsOfEvaluationResponse;
export type ListAuditScopesResponse = SchemaListAuditScopesResponse;
export type ListMetricsResponse = SchemaListMetricsResponse;
export type ListMetricConfigurationsResponse = SchemaListMetricConfigurationResponse;

export function controlUrl(control: Control, catalogId: string): string {
	return `${catalogId}/${control.categoryName}/${control.id}`;
}

export function controlUrl2(cis: ControlInScope, catalogId: string): string {
	return `${catalogId}/${cis.controlCategoryName}/${cis.controlId}`;
}

export function parseControlUrl(url: string): [string, string, string?] {
	const parts = url.split('/');

	return [parts[0], parts[1], parts[2]];
}

// only non-spec type still needed
export interface ListControlsInScopeResponse {
	controlsInScope: ControlInScope[];
}

/**
 * Requests the Clouditor runtime information.
 *
 * @returns an array of {@link AssessmentResult}s.
 */
export async function getRuntimeInfo(fetch = window.fetch): Promise<Runtime> {
	const apiUrl = confirmatize(`/v1/orchestrator/runtime_info`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: Runtime) => {
			return response;
		});
}

/**
 * Requests a list of assessment results from the orchestrator service.
 *
 * @returns an array of {@link AssessmentResult}s.
 */
export async function listAssessmentResults(
	fetch = window.fetch,
	latestByResourceId = false
): Promise<AssessmentResult[]> {
	const apiUrl = confirmatize(
		`/v1/orchestrator/assessment_results?pageSize=1500&latestByResourceId=${latestByResourceId}&asc=false`
	);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListAssessmentResultsResponse) => {
			return response.results;
		});
}

/**
 * Requests a list of assessment results from the orchestrator service.
 *
 * @returns an array of {@link AssessmentResult}s.
 */
export async function listTargetOfEvaluationAssessmentResults(
	serviceId: string,
	fetch = window.fetch
): Promise<AssessmentResult[]> {
	const apiUrl = confirmatize(
		`/v1/orchestrator/assessment_results?pageSize=1000&filter.targetOfEvaluationId=${serviceId}&asc=false`
	);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListAssessmentResultsResponse) => {
			return response.results;
		});
}

/**
 * Retrieves a particular metric from the orchestrator service.
 *
 * @param id the metric id
 * @returns
 */
export async function getMetric(id: string): Promise<Metric> {
	const apiUrl = confirmatize(`/v1/orchestrator/metrics/${id}`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: Metric) => {
			return response;
		});
}

/**
 * Retrieves a particular metric implementation from the orchestrator service.
 *
 * @param id the metric id
 * @returns
 */
export async function getMetricImplementation(id: string): Promise<MetricImplementation> {
	const apiUrl = confirmatize(`/v1/orchestrator/metrics/${id}/implementation`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: MetricImplementation) => {
			return response;
		});
}

/**
 * Retrieves a particular metric configuration from the orchestrator service.
 *
 * @param serviceId the Target of Evaluation ID
 * @param metricId the metric id
 * @returns metric configuration
 */
export async function getMetricConfiguration(
	serviceId: string,
	metricId: string
): Promise<MetricConfiguration> {
	const apiUrl = confirmatize(
		`/v1/orchestrator/targets_of_evaluation/${serviceId}/metric_configurations/${metricId}`
	);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: MetricConfiguration) => {
			return response;
		});
}

/**
 * Retrieves a particular metric implementation from the orchestrator service.
 *
 * @param id the metric id
 * @returns
 */
export async function listMetricConfigurations(
	serviceId: string,
	skipDefault = false
): Promise<Map<string, MetricConfiguration>> {
	const apiUrl = confirmatize(
		`/v1/orchestrator/targets_of_evaluation/${serviceId}/metric_configurations`
	);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListMetricConfigurationsResponse) => {
			const configs = response.configurations ?? {};
			let entries: [string, MetricConfiguration][] = Object.entries(configs);

			if (skipDefault) {
				entries = entries.filter(([, value]) => {
					return !value.isDefault;
				});
			}

			return new Map<string, MetricConfiguration>(entries);
		});
}

/**
 * Creates a new Target of Evaluation
 */
export async function registerTargetOfEvaluation(
	service: TargetOfEvaluation
): Promise<TargetOfEvaluation> {
	const apiUrl = confirmatize(`/v1/orchestrator/targets_of_evaluation`);

	return fetch(apiUrl, {
		method: 'POST',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		},
		body: JSON.stringify(service)
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: TargetOfEvaluation) => {
			return response;
		});
}

/**
 * Removes a Target of Evaluation.
 */
export async function removeTargetOfEvaluation(targetId: string): Promise<void> {
	const apiUrl = confirmatize(`/v1/orchestrator/targets_of_evaluation/${targetId}`);

	return fetch(apiUrl, {
		method: 'DELETE',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json());
}

/**
 * Retrieves a list of Targets of Evaluation from the orchestrator service.
 *
 * @returns an array of {@link TargetOfEvaluation}s.
 */
export async function listTargetsOfEvaluation(fetch = window.fetch): Promise<TargetOfEvaluation[]> {
	const apiUrl = confirmatize(`/v1/orchestrator/targets_of_evaluation`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListTargetsOfEvaluationResponse) => {
			return response.targetsOfEvaluation ?? [];
		});
}

/**
 * Creates a new Audit Scope.
 */
export async function createAuditScope(target: AuditScope): Promise<AuditScope> {
	const apiUrl = confirmatize(`/v1/orchestrator/audit_scopes`);

	return fetch(apiUrl, {
		method: 'POST',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		},
		body: JSON.stringify(target)
	})
		.then(throwError)
		.then((res) => res.json());
}

/**
 * Removes an Audit Scope.
 */
export async function removeAuditScope(target: AuditScope): Promise<AuditScope> {
	const apiUrl = confirmatize(`/v1/orchestrator/audit_scopes/${target.id}`);

	return fetch(apiUrl, {
		method: 'DELETE',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		},
		body: JSON.stringify(target)
	})
		.then(throwError)
		.then((res) => res.json());
}

/**
 * Retrieves a list of Audit Scopes from the orchestrator service.
 *
 * @returns an array of {@link AuditScope}s.
 */
export async function listAuditScopes(
	serviceId: string,
	fetch = window.fetch
): Promise<AuditScope[]> {
	const apiUrl = confirmatize(
		`/v1/orchestrator/audit_scopes?filter.targetOfEvaluationId=${serviceId}`
	);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListAuditScopesResponse) => {
			return response.auditScopes ?? [];
		});
}

/**
 * Retrieves a list of catalogs from the orchestrator service.
 *
 * @returns an array of {@link Catalog}s.
 */
export async function listCatalogs(fetch = window.fetch): Promise<Catalog[]> {
	const apiUrl = confirmatize(`/v1/orchestrator/catalogs`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListCatalogsResponse) => {
			return response.catalogs ?? [];
		});
}

/**
 * Retrieves a catalog from the orchestrator service.
 *
 * @returns a {@link Catalog}s.
 */
export async function getCatalog(id: string, fetch = window.fetch): Promise<Catalog> {
	const apiUrl = confirmatize(`/v1/orchestrator/catalogs/${id}`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json());
}

/**
 * Retrieves controls from the orchestrator service.
 *
 * @returns a list of {@link Control}s.
 */
export async function listControls(
	catalogId?: string,
	categoryName?: string,
	fetch = window.fetch
): Promise<Control[]> {
	let baseApiUrl: string;
	if (catalogId != null && categoryName != null) {
		baseApiUrl = confirmatize(
			`/v1/orchestrator/catalogs/${catalogId}/categories/${categoryName}/controls?pageSize=1500&orderBy=id&asc=true`
		);
	} else {
		baseApiUrl = confirmatize(`/v1/orchestrator/controls?pageSize=1500&orderBy=id&asc=true`);
	}

	const controls: Control[] = [];
	let nextPageToken = undefined;
	do {
		let apiUrl = baseApiUrl;
		if (nextPageToken != undefined) {
			apiUrl = apiUrl + '&pageToken=' + nextPageToken;
		}

		const res = (await throwError(
			await fetch(apiUrl, {
				method: 'GET',
				headers: {
					Authorization: `Bearer ${localStorage.token}`
				}
			})
		).json()) as ListControlsResponse;

		controls.push(...(res.controls ?? []));
		nextPageToken = res.nextPageToken;
	} while (nextPageToken != '' && nextPageToken !== undefined && nextPageToken !== null);

	return controls;
}

/**
 * Retrieve a Target of Evaluation from the orchestrator service using its ID.
 *
 * @returns the Target of Evaluation
 */
export async function getTargetOfEvaluation(
	id: string,
	fetch = window.fetch
): Promise<TargetOfEvaluation> {
	const apiUrl = confirmatize(`/v1/orchestrator/targets_of_evaluation/${id}`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json());
}

export async function updateTargetOfEvaluation(
	service: TargetOfEvaluation,
	fetch = window.fetch
): Promise<TargetOfEvaluation> {
	const apiUrl = confirmatize(`/v1/orchestrator/targets_of_evaluation/${service.id}`);

	return fetch(apiUrl, {
		method: 'PUT',
		body: JSON.stringify(service),
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: TargetOfEvaluation) => {
			return response;
		});
}

export async function updateControlInScope(
	scope: ControlInScope,
	fetch = window.fetch
): Promise<TargetOfEvaluation> {
	const apiUrl = confirmatize(
		`/v1/orchestrator/targets_of_evaluation/${scope.auditScopeTargetOfEvaluationId}/audit_scopes/${scope.auditScopeCatalogId}/controls_in_scope/categories/${scope.controlCategoryName}/controls/${scope.controlId}`
	);

	return fetch(apiUrl, {
		method: 'PUT',
		body: JSON.stringify(scope),
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: TargetOfEvaluation) => {
			return response;
		});
}

/**
 * Retrieves a list of metrics from the orchestrator service.
 *
 * @returns an array of {@link Metric}s.
 */
export async function listMetrics(fetch = window.fetch): Promise<Metric[]> {
	const apiUrl = confirmatize(`/v1/orchestrator/metrics?pageSize=200`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListMetricsResponse) => {
			return response.metrics ?? [];
		});
}

/**
 * Retrieves a list of certificates from the orchestrator service.
 *
 * @returns an array of {@link Certificate}s.
 */
export async function listCertificates(): Promise<Certificate[]> {
	const apiUrl = confirmatize(`/v1/orchestrator/certificates`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListCertificatesResponse) => {
			return response.certificates ?? [];
		});
}
