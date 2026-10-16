import type { Connectivity, OperatorSkill, PhysicalAccess, RiskProfile, RiskScenario } from './profile';

/** The scale of the impact on an asset. 0 means that the property does not matter for the asset. */
export const impactLevels = [
	{ value: '0', label: 'Not relevant' },
	{ value: '1', label: '1, very low' },
	{ value: '2', label: '2, low' },
	{ value: '3', label: '3, moderate' },
	{ value: '4', label: '4, high' },
	{ value: '5', label: '5, very high' }
];

const impactNames = ['not relevant', 'very low', 'low', 'moderate', 'high', 'very high'];

/** The impact level in words, for example "high (4)". */
export function impactName(level: number): string {
	return level > 0 ? `${impactNames[level] ?? level} (${level})` : impactNames[0];
}

export const physicalAccessOptions: { value: PhysicalAccess; label: string; hint: string }[] = [
	{ value: 'CONTROLLED', label: 'Controlled', hint: 'Only authorised people can reach the device or the system it runs on' },
	{ value: 'SUPERVISED', label: 'Supervised', hint: 'Many people can reach it, but it is watched' },
	{ value: 'OPEN', label: 'Open', hint: 'Anybody can reach it, or it is carried around' }
];

export const connectivityOptions: { value: Connectivity; label: string; hint: string }[] = [
	{ value: 'LOCAL', label: 'Local only', hint: 'No network, only local interfaces' },
	{ value: 'PRIVATE_NETWORK', label: 'Private network', hint: 'A network with known and trusted participants' },
	{ value: 'INTERNET', label: 'Internet', hint: 'It can be reached from the internet' }
];

export const operatorSkillOptions: { value: OperatorSkill; label: string; hint: string }[] = [
	{ value: 'EXPERT', label: 'Trained administrators', hint: 'People who know how to secure it' },
	{ value: 'GENERAL', label: 'General users', hint: 'People without security knowledge, for example consumers' }
];

export function emptyRisk(): RiskProfile {
	return { assets: [], physicalAccess: '', connectivity: '', operatorSkill: '' };
}

/** Whether the value of the profile is one of the allowed values. Not said, or no restriction, allows it. */
function envReached<T extends string>(value: T | '', allowed: T[]): boolean {
	return allowed.length === 0 || value === '' || allowed.includes(value);
}

/**
 * Whether the risk profile reaches the scenario of a requirement, so that the requirement applies. This is the same
 * rule as the one of the server: some asset has at least the impacts of the scenario at the same time, and the
 * environment is one of the given ones. If the profile does not say something that the scenario depends on, the
 * requirement applies: it is only left out if the profile shows that the risk is lower.
 */
export function scenarioReached(profile: RiskProfile, scenario: RiskScenario | undefined): boolean {
	if (!scenario) return true;

	if (
		!envReached(profile.physicalAccess, scenario.physicalAccess) ||
		!envReached(profile.connectivity, scenario.connectivity) ||
		!envReached(profile.operatorSkill, scenario.operatorSkill)
	) {
		return false;
	}

	if (scenario.minConfidentiality === 0 && scenario.minIntegrity === 0 && scenario.minAvailability === 0) return true;
	if (profile.assets.length === 0) return true;

	return profile.assets.some(
		(a) =>
			a.confidentiality >= scenario.minConfidentiality &&
			a.integrity >= scenario.minIntegrity &&
			a.availability >= scenario.minAvailability
	);
}

function names<T extends string>(values: T[], options: { value: T; label: string }[]): string {
	return values.map((v) => options.find((o) => o.value === v)?.label.toLowerCase() ?? v.toLowerCase()).join(' or ');
}

/** Describes in plain words the risk at which a requirement applies. */
export function describeScenario(s: RiskScenario): string {
	const impacts = [
		s.minConfidentiality > 0 ? `confidentiality ${impactNames[s.minConfidentiality]} (${s.minConfidentiality}) or higher` : '',
		s.minIntegrity > 0 ? `integrity ${impactNames[s.minIntegrity]} (${s.minIntegrity}) or higher` : '',
		s.minAvailability > 0 ? `availability ${impactNames[s.minAvailability]} (${s.minAvailability}) or higher` : ''
	].filter(Boolean);
	const env = [
		s.physicalAccess.length ? `physical access ${names(s.physicalAccess, physicalAccessOptions)}` : '',
		s.connectivity.length ? `connectivity ${names(s.connectivity, connectivityOptions)}` : '',
		s.operatorSkill.length ? `operated by ${names(s.operatorSkill, operatorSkillOptions)}` : ''
	].filter(Boolean);

	const parts = [
		impacts.length ? `an asset with ${impacts.join(' and ')}` : '',
		env.length ? `in an environment with ${env.join(', ')}` : ''
	].filter(Boolean);
	return parts.length ? `Applies for ${parts.join(' ')}.` : 'Applies at any risk.';
}
