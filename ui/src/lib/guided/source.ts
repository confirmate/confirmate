import { mockSource } from './mock';

/** What the list of targets of evaluation shows about one target. */
export interface TargetSummary {
	id: string;
	name: string;
	subtitle: string;
	typeLabel: string;
	frameworkLabel: string;
	tags: string[];
	progress: Progress;
	nextDeadline?: string;
}

export interface Progress {
	met: number;
	attention: number;
	open: number;
	total: number;
}

export type RequirementStatus = 'met' | 'attention' | 'open';

/** The progress of one part of the framework, for example "Vulnerability handling". */
export interface PartProgress extends Progress {
	id: string;
	label: string;
}

export interface Task {
	id: string;
	title: string;
	detail: string;
	urgency: 'urgent' | 'attention' | 'open';
	assignee: string;
	due?: string;
	/** The requirement the task belongs to, if it has a view of its own. */
	requirementId?: string;
}

/** A running deadline such as the reporting clock of an exploited vulnerability. */
export interface Clock {
	title: string;
	text: string;
	action: string;
	steps: { label: string; due: string }[];
}

export interface SourceStatus {
	name: string;
	state: 'ok' | 'partial' | 'none';
	detail: string;
}

export interface RequirementSummary {
	id: string;
	partId: string;
	title: string;
	/** The place in the regulation or standard. */
	reference: string;
	status: RequirementStatus;
}

export interface Check {
	id: string;
	title: string;
	/** Automatic checks come from a source, statements are written by a person. */
	kind: 'automatic' | 'statement';
	status: 'met' | 'not-met' | 'open';
	text: string;
	code?: string;
	note?: string;
	actions: { label: string; primary?: boolean }[];
}

export interface RequirementDetail extends RequirementSummary {
	plain: string;
	scopeReason: string;
	checks: Check[];
	history: { when: string; what: string }[];
	comments: { who: string; text: string }[];
}

export interface TargetDetail extends TargetSummary {
	chips: string[];
	clock?: Clock;
	tasks: Task[];
	parts: PartProgress[];
	sources: SourceStatus[];
	requirements: RequirementSummary[];
	/** The requirements that have a view of their own in the preview. */
	requirementDetails: RequirementDetail[];
}

/**
 * Where the guided experience gets its data from. Today this is a mock; the real source reads targets of evaluation,
 * evaluation results and deadlines from the orchestrator. The screens only depend on this interface.
 */
export interface GuidedSource {
	listTargets(): Promise<TargetSummary[]>;
	getTarget(id: string): Promise<TargetDetail | undefined>;
}

export const guidedSource: GuidedSource = mockSource;
