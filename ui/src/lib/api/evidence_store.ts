import { throwError } from './errors';
import { confirmatize } from './util';

export interface GraphEdge {
	id: string;
	source: string;
	target: string;
	type: string;
}

export interface ListGraphEdgesRequest {
	orderBy?: string;
	pageSize?: number;
	asc?: boolean;
}
export interface ListGraphEdgesResponse {
	edges: GraphEdge[];
}

export interface ListResourcesRequest {
	filter?: Filter;
	orderBy?: string;
}

export interface ListResourcesResponse {
	results: Resource[];
}

export interface Filter {
	targetOfEvaluationId?: string;
	type?: string;
}

export interface ResourceProperties {
	name: string;
	raw: string;
	type: string[];
	serviceId: string;
	id: string;
	labels: object;
}

export interface Resource {
	id: string;
	targetOfEvaluationId: string;
	resourceType: string;
	properties: ResourceProperties;
}

export async function listResources(
	filteredTargetOfEvaluationId?: string,
	type = '',
	fetch = window.fetch
): Promise<Resource[]> {
	let apiUrl =
		confirmatize(`/v1/evidence_store/resources?filter.targetOfEvaluationId=${filteredTargetOfEvaluationId}&pageSize=1500&orderBy=resource_type&asc=true
    `);

	if (type != '') {
		apiUrl += `&filter.type=${type}`;
	}

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListResourcesResponse) => {
			return response.results;
		});
}

export async function listGraphEdges(fetch = window.fetch): Promise<GraphEdge[]> {
	const apiUrl = confirmatize(`/v1experimental/evidence/graph/edges`); //?pageSize=1500`);

	return fetch(apiUrl, {
		method: 'GET',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then(throwError)
		.then((res) => res.json())
		.then((response: ListGraphEdgesResponse) => {
			return response.edges;
		});
}
