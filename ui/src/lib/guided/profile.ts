/**
 * A framework profile describes how the guided experience works for one compliance framework. It is plain data, so
 * that the same generic components serve the CRA, BSI C5, NIS-2 and others. Later the profile is delivered by the
 * backend together with the catalog; for now it is a TypeScript object.
 */

export type Answer = 'yes' | 'no' | 'unknown';

/** The kinds of steps the wizard knows. A profile picks and orders them. */
export type StepType = 'details' | 'choice' | 'risk' | 'scope' | 'sources';

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

/** How a requirement applies: always, if a condition holds, or depending on the risk of the product. */
export type ApplicabilityKind = 'always' | 'conditional' | 'risk';

/** What kind of evidence a requirement asks for. */
export type RequirementType = 'technical' | 'activity' | 'documentation';

export type PhysicalAccess = 'CONTROLLED' | 'SUPERVISED' | 'OPEN';
export type Connectivity = 'LOCAL' | 'PRIVATE_NETWORK' | 'INTERNET';
export type OperatorSkill = 'EXPERT' | 'GENERAL';

/** The risk at which a requirement applies. A value that is not set does not restrict. */
export interface RiskScenario {
	/** The lowest impact on an asset, from 1 (very low) to 5 (very high). 0 means that it does not matter. */
	minConfidentiality: number;
	minIntegrity: number;
	minAvailability: number;
	/** The environments in which the requirement applies. An empty list means any. */
	physicalAccess: PhysicalAccess[];
	connectivity: Connectivity[];
	operatorSkill: OperatorSkill[];
}

export interface Applicability {
	kind: ApplicabilityKind;
	/** For "conditional": the statement that has to be true for the requirement to apply. */
	condition?: string;
	/** For "risk": the risk at which the requirement applies. */
	scenario?: RiskScenario;
	/** The short names of requirements that can be fulfilled instead, if they give at least the same protection. */
	compensatedBy: string[];
}

/** A place where a requirement comes from, for example the legal basis. */
export interface Reference {
	framework: string;
	identifier: string;
	relationship: string;
	url?: string;
}

/** A single requirement that can be taken out of scope, as an atomic, assessable obligation. */
export interface ScopableRequirement {
	/** The short name of the control in the catalog. */
	id: string;
	title: string;
	text: string;
	/** Number of checks (metrics) behind the requirement. */
	checks: number;
	type: RequirementType;
	applicability: Applicability;
	references: Reference[];
}

/** A group of requirements, for example the requirements behind one point of the legal text. */
export interface RequirementGroup {
	/** The short name of the control that groups the requirements. */
	id: string;
	title: string;
	/** The place in the law or standard, for example "Annex I, Part I, (2)(e)". */
	reference: string;
	/** The part of the framework, for example "Product security requirements (Annex I, Part I)". */
	part: string;
	requirements: ScopableRequirement[];
}

/** The impact on one asset if its confidentiality, integrity or availability is lost. 0 means not relevant. */
export interface AssetImpact {
	name: string;
	confidentiality: number;
	integrity: number;
	availability: number;
}

/** The part of a risk assessment that tells which requirements apply: the assets and the environment. */
export interface RiskProfile {
	assets: AssetImpact[];
	/** Empty if not said. */
	physicalAccess: PhysicalAccess | '';
	connectivity: Connectivity | '';
	operatorSkill: OperatorSkill | '';
}

/** A ready-made starting point for the risk profile, which a person adjusts. */
export interface RiskPreset {
	id: string;
	label: string;
	hint: string;
	profile: RiskProfile;
}

/** What the risk step of the wizard offers. */
export interface RiskStepContent {
	heading: string;
	presets: RiskPreset[];
	/** Names offered for new assets. */
	assetSuggestions: string[];
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
	risk: RiskStepContent;
	sources: SourceCard[];
}
