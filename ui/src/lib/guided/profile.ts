/**
 * A framework profile describes how the guided experience works for one compliance framework. It is plain data, so
 * that the same generic components serve the CRA, BSI C5, NIS-2 and others. Later the profile is delivered by the
 * backend together with the catalog; for now it is a TypeScript object.
 */

export type Answer = 'yes' | 'no' | 'unknown';

/** The kinds of steps the wizard knows. A profile picks and orders them. */
export type StepType = 'details' | 'choice' | 'questionnaire' | 'scope' | 'sources';

export interface ProfileStep {
	id: string;
	type: StepType;
	label: string;
	title: string;
	intro: string;
}

export interface ChoiceOption {
	id: string;
	label: string;
	hint: string;
	/** What the choice sets on the target of evaluation, for example its classification. */
	attributes?: Record<string, string>;
	/** What the choice means, shown next to the options. */
	result: { title: string; text: string; note?: string };
}

export interface Question {
	id: string;
	text: string;
	help?: string;
	defaultAnswer: Answer;
}

/** How a requirement is applied: always, only if the product has a capability, or depending on the risk. */
export type Applicability = 'mandatory' | 'capability' | 'risk';

/** A single requirement that can be taken out of scope, as an atomic, assessable obligation. */
export interface ScopableRequirement {
	/** The short name of the control in the catalog. */
	id: string;
	title: string;
	text: string;
	/** Number of checks (metrics) behind the requirement. */
	checks: number;
	applicability: Applicability;
	/** For applicability "capability": the requirement only applies if the answer to the question is yes. */
	capability?: { question: string; text: string };
}

/** A group of requirements, for example the requirements behind one point of the legal text. */
export interface RequirementGroup {
	/** The short name of the control that groups the requirements. */
	id: string;
	title: string;
	reference: string;
	requirements: ScopableRequirement[];
}

/**
 * A suggestion for the scope that follows from an answer. If every condition is an explicit match, the requirements are
 * suggested as not applicable (or as an accepted residual risk), with the reason. A person decides; "don't know" never
 * matches.
 */
export interface ScopeSuggestion {
	when: Partial<Record<string, Answer>>;
	requirements: string[];
	treatment: 'na' | 'risk';
	reason: string;
}

export interface QuestionNote {
	when: Partial<Record<string, Answer>>;
	tone: 'warn' | 'brand';
	text: string;
}

export interface Questionnaire {
	heading: string;
	questions: Question[];
	groups: RequirementGroup[];
	suggestions: ScopeSuggestion[];
	notes: QuestionNote[];
}

export interface DetailsField {
	id: string;
	label: string;
	hint?: string;
	type: 'text' | 'date' | 'select';
	options?: string[];
	defaultValue: string;
}

/** A place where evidence comes from, as shown in the sources step. */
export interface SourceCard {
	id: string;
	title: string;
	text: string;
	/** Sample state until the real connections exist. */
	state: 'connected' | 'partial' | 'optional';
	stateLabel: string;
	detail?: string;
	actions?: string[];
	snippet?: string;
}

export interface FrameworkProfile {
	id: string;
	/** Short name of the framework as shown to users. */
	name: string;
	/** The catalog that the audit scope of a target of evaluation is created for. */
	catalogId: string;
	/** Type of target of evaluation this profile is made for (TargetOfEvaluation.TargetType of the API). */
	targetType: 'TARGET_TYPE_PRODUCT' | 'TARGET_TYPE_CLOUD' | 'TARGET_TYPE_ORGANIZATION';
	typeLabel: string;
	/** What the person wants to check, shown on the first screen. */
	pitch: { title: string; text: string };
	steps: ProfileStep[];
	details: DetailsField[];
	/** The choice step, for example the product class. */
	choices: ChoiceOption[];
	questionnaire: Questionnaire;
	sources: SourceCard[];
	/** Number of requirements that always apply, regardless of the questionnaire. */
	alwaysApplies: { count: number; text: string };
}
