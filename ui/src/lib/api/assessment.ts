export interface AssessmentResult {
	id: string;
	timestamp: string;
	metricId: string;
	metricConfiguration: MetricConfiguration;
	evidenceId: string;
	compliant: boolean;
	resourceId: string;
	resourceTypes: string[];
	nonComplianceComments: string;
}

export interface MetricConfiguration {
	operator: string;
	targetValue: object;
	isDefault: boolean;
}

export interface Metric {
	id: string;
	name: string;
	description: string;
	category: string;
}

export interface MetricImplementation {
	metricId: string;
	lang: string;
	code: string;
}
