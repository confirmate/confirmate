import type { Answer, FrameworkProfile, RequirementGroup, RiskProfile } from './profile';
import type { Decisions, NotApplicableReason } from './scope';

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
	/** The metric behind the check, if a person can answer it with a statement. */
	metricId?: string;
	actions: { id: 'statement' | 'other'; label: string; primary?: boolean }[];
}

export interface RequirementDetail extends RequirementSummary {
	plain: string;
	/** The short names of requirements that can be fulfilled instead of this one. */
	compensatedBy: string[];
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
}

/** What the setup wizard collected about a new Target of Evaluation. */
export interface NewTarget {
	profile: FrameworkProfile;
	details: Record<string, string>;
	choiceId: string;
	/** The requirements of the catalog that the scope was worked out for. */
	groups: RequirementGroup[];
	/** The assets and the environment of the risk assessment. */
	risk: RiskProfile;
	/** The answers about the conditions of the requirements, by condition ID. */
	answers: Record<string, Answer>;
	/** The decisions about single requirements that go with the answers. */
	decisions: Decisions;
}

/** A statement that a person makes about a check, for example that a requirement is met. */
export interface NewStatement {
	targetId: string;
	metricId: string;
	compliant: boolean;
	comment: string;
	/** Optional day until which the statement is valid (YYYY-MM-DD). */
	validUntil?: string;
}

/** A residual risk that a person accepted for a requirement. */
export interface AcceptedRisk {
	id: string;
	requirement: string;
	reference: string;
	text: string;
	acceptedBy: string;
	acceptedAt: string;
	reviewBy: string;
}

/** A change of the treatment of one requirement after the setup. */
export interface ScopeChange {
	targetId: string;
	requirementId: string;
	treatment: 'na' | 'risk';
	reason: string;
	/** For "not applicable": why. */
	naReason?: NotApplicableReason;
	/** For the reason "compensated": the control that is fulfilled instead. */
	compensatedBy?: string;
}

/** One decision of a risk assessment about a requirement, as it is shown after the setup. */
export interface AssessmentDecision {
	requirementId: string;
	title: string;
	reference: string;
	/** Taken out of the scope, kept although the risk profile would take it out, or kept with an accepted risk. */
	outcome: 'not-applicable' | 'accepted-risk' | 'kept';
	/** For "not applicable": why, as one of the fixed reasons. */
	reason?: NotApplicableReason;
	/** The requirement that is fulfilled instead, for the reason "compensated". */
	compensatedBy?: string;
	text: string;
	/** Where the decision comes from: the rules of the risk profile, an import or a person. */
	source: 'rule' | 'import' | 'person';
}

export interface AssessmentVersion {
	version: number;
	status: 'draft' | 'approved' | 'superseded';
	createdAt: string;
	summary: string;
}

/** The approved risk assessment of a Target of Evaluation and what it changed. */
export interface RiskAssessmentView {
	version: number;
	method: string;
	summary: string;
	approvedBy: string;
	approvedAt: string;
	reviewDueAt: string;
	/** The assets and the environment, if the assessment was made from a risk profile. */
	profile?: RiskProfile;
	decisions: AssessmentDecision[];
	versions: AssessmentVersion[];
}

/**
 * Where the guided experience gets its data from. There is a source with sample data and one that reads from the
 * orchestrator. The screens only depend on this interface.
 */
export interface GuidedSource {
	listTargets(): Promise<TargetSummary[]>;
	getTarget(id: string): Promise<TargetDetail | undefined>;
	getRequirement(targetId: string, requirementId: string): Promise<RequirementDetail | undefined>;
	/** The requirements of a catalog with their legal basis and when they apply, grouped by point of the law. */
	listRequirementGroups(catalogId: string): Promise<RequirementGroup[]>;
	/** Creates the Target of Evaluation, its audit scope and an approved risk assessment, and returns its ID. */
	createTarget(input: NewTarget): Promise<{ id: string }>;
	addStatement(statement: NewStatement): Promise<void>;
	/** Takes a requirement out of scope or accepts its residual risk, as a new version of the risk assessment. */
	changeScope(change: ScopeChange): Promise<void>;
	listAcceptedRisks(targetId: string): Promise<AcceptedRisk[]>;
	/** The approved risk assessment of the target and what it changed, or nothing if there is none. */
	getRiskAssessment(targetId: string): Promise<RiskAssessmentView | undefined>;
}
