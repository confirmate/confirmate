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
	/** What the choice means, shown next to the options. */
	result: { title: string; text: string; note?: string };
}

export interface Question {
	id: string;
	text: string;
	help?: string;
	defaultAnswer: Answer;
}

/** The requirements that a questionnaire can switch out of scope. */
export interface ScopableRequirement {
	id: string;
	title: string;
}

/** If all conditions hold, the requirements are out of scope, with the reason. Unknown answers never match "no". */
export interface ScopeEffect {
	when: Partial<Record<string, Answer>>;
	out: string[];
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
	requirements: ScopableRequirement[];
	effects: ScopeEffect[];
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
