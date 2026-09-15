import type {
	SchemaCategory,
	SchemaControl,
	SchemaControlInScope,
	SchemaEvaluationResult
} from '$lib/api/openapi/orchestrator';

/**
 * Buckets top-level controls by the name of the category they belong to, using the catalog's
 * nested category/control structure to look up each control's category.
 */
export function bucketControlsByCategory(
	categories: SchemaCategory[],
	topLevelControls: SchemaControl[]
): Record<string, SchemaControl[]> {
	const categoryByControlId = new Map<string, string>();
	for (const cat of categories) {
		for (const c of cat.controls ?? []) {
			if (c.id) categoryByControlId.set(c.id, cat.name ?? '');
		}
	}

	return Object.fromEntries(
		categories.map((cat) => [
			cat.name,
			topLevelControls.filter((c) => categoryByControlId.get(c.id ?? '') === cat.name)
		])
	);
}

/** Recursively flattens a control tree (including sub-controls) into a single array. */
export function flattenControls(controls: SchemaControl[]): SchemaControl[] {
	return controls.flatMap((c) => [c, ...flattenControls(c.controls ?? [])]);
}

/** Builds a lookup of control ID to short name across a control tree, for audit trail display. */
export function indexControlsById(
	controls: SchemaControl[]
): Record<string, { shortName?: string }> {
	return Object.fromEntries(
		flattenControls(controls)
			.filter((c) => c.id)
			.map((c) => [c.id!, { shortName: c.shortName }])
	);
}

/** Indexes evaluation results by control ID, keeping the last result seen per control. */
export function indexEvaluationsByControl(
	results: SchemaEvaluationResult[]
): Record<string, SchemaEvaluationResult> {
	const byControl: Record<string, SchemaEvaluationResult> = {};
	for (const result of results) {
		const key = result.controlId ?? '';
		if (key) byControl[key] = result;
	}
	return byControl;
}

/** Indexes ControlInScope records by control ID. */
export function indexControlsInScope(
	records: SchemaControlInScope[]
): Record<string, SchemaControlInScope> {
	const byControlId: Record<string, SchemaControlInScope> = {};
	for (const cis of records) {
		if (cis.controlId) byControlId[cis.controlId] = cis;
	}
	return byControlId;
}

interface AssessmentResultLike {
	id?: string;
	metricId?: string;
	compliant?: boolean;
	createdAt?: string;
}

/** Indexes assessment results by their ID, keeping only the display-relevant fields. */
export function indexAssessmentsById(
	results: AssessmentResultLike[]
): Record<string, { metricId?: string; compliant?: boolean; createdAt?: string }> {
	const byId: Record<string, { metricId?: string; compliant?: boolean; createdAt?: string }> = {};
	for (const ar of results) {
		if (ar.id) byId[ar.id] = { metricId: ar.metricId, compliant: ar.compliant, createdAt: ar.createdAt };
	}
	return byId;
}

/** Counts passing/failing assessment results per metric ID. */
export function countAssessmentsByMetric(
	results: AssessmentResultLike[]
): Record<string, { passing: number; failing: number }> {
	const countByMetric: Record<string, { passing: number; failing: number }> = {};
	for (const ar of results) {
		const metricName = ar.metricId;
		if (!metricName) continue;
		countByMetric[metricName] ??= { passing: 0, failing: 0 };
		if (ar.compliant) {
			countByMetric[metricName].passing++;
		} else {
			countByMetric[metricName].failing++;
		}
	}
	return countByMetric;
}
