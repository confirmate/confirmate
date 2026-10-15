import { describe, expect, it } from 'vitest';
import catalog from '../../../../example-data/catalogs/cra-example-catalog.json';
import { conditionId, conditionsOf, groupsOf } from './catalog';
import type { Applicability, RequirementGroup, RiskProfile, RiskScenario, ScopableRequirement } from './profile';
import { emptyRisk, scenarioReached } from './risk';
import { computeScope, summarize } from './scope';

const high: RiskScenario = {
	minConfidentiality: 4,
	minIntegrity: 0,
	minAvailability: 0,
	physicalAccess: [],
	connectivity: ['INTERNET'],
	operatorSkill: []
};

const risk = (partial: Partial<RiskProfile>): RiskProfile => ({ ...emptyRisk(), ...partial });
const asset = (confidentiality = 0, integrity = 0, availability = 0) => ({ name: 'Asset', confidentiality, integrity, availability });

function requirement(id: string, applicability: Partial<Applicability>): ScopableRequirement {
	return {
		id,
		title: id,
		text: '',
		checks: 1,
		type: 'technical',
		applicability: { kind: 'always', compensatedBy: [], ...applicability },
		references: []
	};
}

const group = (...requirements: ScopableRequirement[]): RequirementGroup => ({
	id: 'point',
	title: 'A point',
	reference: 'Annex I',
	part: 'Part',
	requirements
});

describe('scenarioReached', () => {
	it.each([
		['no scenario', risk({}), undefined, true],
		['impact and environment reached', risk({ assets: [asset(5)], connectivity: 'INTERNET' }), high, true],
		['impact too low', risk({ assets: [asset(2)], connectivity: 'INTERNET' }), high, false],
		['environment not reached', risk({ assets: [asset(5)], connectivity: 'LOCAL' }), high, false],
		[
			'the impacts have to be reached by one asset',
			risk({ assets: [asset(5), asset(0, 5)] }),
			{ ...high, minIntegrity: 4, connectivity: [] },
			false
		],
		['environment not given in the profile', risk({ assets: [asset(5)] }), high, true],
		['no assets in the profile', risk({ connectivity: 'INTERNET' }), high, true]
	])('%s', (_name, profile, scenario, want) => {
		expect(scenarioReached(profile, scenario)).toBe(want);
	});
});

describe('computeScope', () => {
	const always = requirement('always', { kind: 'always' });
	const conditional = requirement('conditional', { kind: 'conditional', condition: 'The product has accounts' });
	const riskBased = requirement('risk', { kind: 'risk', scenario: high });
	const answers = (a: 'yes' | 'no' | 'unknown') => ({ [conditionId('The product has accounts')]: a });
	const rowOf = (id: string, groups: ReturnType<typeof computeScope>) => groups[0].rows.find((r) => r.requirement.id === id)!;

	it('a requirement that always applies stays and cannot be decided', () => {
		const rows = computeScope([group(always, conditional)], risk({}), answers('no'), {
			decided: { always: { treatment: 'na', reason: 'x' } }
		});
		expect(rowOf('always', rows)).toMatchObject({ treatment: 'applies', locked: true });
	});

	it.each([
		['yes', 'applies'],
		['unknown', 'applies'],
		['no', 'na']
	] as const)('a conditional requirement with the answer %s is %s', (answer, want) => {
		const rows = computeScope([group(always, conditional)], risk({}), answers(answer), { decided: {} });
		expect(rowOf('conditional', rows).treatment).toBe(want);
		if (want === 'na') expect(rowOf('conditional', rows)).toMatchObject({ naReason: 'CONDITION_NOT_MET', source: 'rule' });
	});

	it('a risk based requirement is not applicable below its scenario', () => {
		const rows = computeScope([group(always, riskBased)], risk({ assets: [asset(1)], connectivity: 'LOCAL' }), {}, { decided: {} });
		expect(rowOf('risk', rows)).toMatchObject({ treatment: 'na', naReason: 'BELOW_RISK_LEVEL', source: 'rule' });
		expect(rowOf('risk', rows).reason).toContain('Your risk profile is below that');
	});

	it('a risk based requirement applies if the scenario is reached', () => {
		const rows = computeScope([group(always, riskBased)], risk({ assets: [asset(5)], connectivity: 'INTERNET' }), {}, { decided: {} });
		expect(rowOf('risk', rows)).toMatchObject({ treatment: 'applies', source: 'default' });
	});

	it('risk based requirements stay if the rules would take out the whole point', () => {
		const rows = computeScope([group(riskBased, requirement('risk2', { kind: 'risk', scenario: high }))], risk({ assets: [asset(1)], connectivity: 'LOCAL' }), {}, { decided: {} });
		expect(rows[0].rows.map((r) => r.treatment)).toEqual(['applies', 'applies']);
		expect(rows[0].empty).toBe(false);
	});

	it('a point is empty if the answers take out all its conditional requirements', () => {
		const rows = computeScope([group(conditional)], risk({}), answers('no'), { decided: {} });
		expect(rows[0].empty).toBe(true);
		expect(summarize(rows).emptyGroups).toBe(1);
	});

	it('a decision of a person wins over the rules and keeps a requirement', () => {
		const rows = computeScope([group(always, riskBased)], risk({ assets: [asset(1)], connectivity: 'LOCAL' }), {}, {
			decided: { risk: { treatment: 'applies', reason: '' } }
		});
		expect(rowOf('risk', rows)).toMatchObject({ treatment: 'applies', source: 'person' });
	});

	it('the last requirement of a point cannot be taken out', () => {
		const other = requirement('other', { kind: 'conditional', condition: 'The product has a debug interface' });
		const rows = computeScope([group(conditional, other)], risk({}), { ...answers('yes'), [conditionId('The product has a debug interface')]: 'no' }, { decided: {} });
		expect(rowOf('conditional', rows).naBlocked).toBe(true);
		expect(rowOf('other', rows).naBlocked).toBe(false);
	});

	it('a reason is required, and a compensated requirement needs the one that replaces it', () => {
		const rows = computeScope([group(always, conditional, riskBased)], risk({ assets: [asset(5)], connectivity: 'INTERNET' }), answers('yes'), {
			decided: {
				conditional: { treatment: 'na', reason: '' },
				risk: { treatment: 'na', reason: 'Another one is used', naReason: 'COMPENSATED' }
			}
		});
		expect(rowOf('conditional', rows).missingReason).toBe(true);
		expect(rowOf('risk', rows).missingCompensation).toBe(true);
		expect(summarize(rows).review).toBe(2);
	});
});

describe('the CRA example catalog', () => {
	const groups = groupsOf([{ name: 'CRA', controls: catalog[0].categories.flatMap((c) => c.controls) as never }]);
	const requirements = groups.flatMap((g) => g.requirements);

	it('has requirements with a legal basis, a type and an applicability', () => {
		expect(requirements.length).toBeGreaterThan(60);
		for (const r of requirements) {
			expect(r.references.some((x) => x.framework === 'CRA' && x.relationship === 'legal-basis')).toBe(true);
			expect(['technical', 'activity', 'documentation']).toContain(r.type);
		}
		for (const g of groups) expect(g.reference).toMatch(/^(Annex|Art)/);
	});

	it('names the requirements that can compensate a requirement', () => {
		const ids = new Set(requirements.map((r) => r.id));
		for (const r of requirements) for (const c of r.applicability.compensatedBy) expect(ids.has(c)).toBe(true);
	});

	it('every risk based requirement has a scenario and every conditional one a condition', () => {
		for (const r of requirements) {
			if (r.applicability.kind === 'risk') expect(r.applicability.scenario).toBeDefined();
			if (r.applicability.kind === 'conditional') expect(r.applicability.condition).toBeTruthy();
		}
		expect(conditionsOf(groups).length).toBeGreaterThan(5);
	});

	it('takes out the same requirements as the server for the same risk profile', () => {
		// The server derived 12 exclusions for this profile when it was run against the catalog (reason: below the risk
		// level). Conditions do not play a part, so the answers are left out.
		const profile = risk({ assets: [asset(3, 2, 1)], physicalAccess: 'CONTROLLED', connectivity: 'LOCAL', operatorSkill: 'GENERAL' });
		const scoped = computeScope(groups, profile, {}, { decided: {} });
		const out = scoped.flatMap((g) => g.rows).filter((r) => r.treatment === 'na' && r.naReason === 'BELOW_RISK_LEVEL');
		expect(out).toHaveLength(12);
	});

	it('a local desktop tool keeps far fewer requirements than an exposed internet device', () => {
		const desktop = risk({ assets: [asset(3, 2, 1)], physicalAccess: 'CONTROLLED', connectivity: 'LOCAL' });
		const device = risk({ assets: [asset(5, 5, 4)], physicalAccess: 'OPEN', connectivity: 'INTERNET' });

		const small = summarize(computeScope(groups, desktop, {}, { decided: {} }));
		const big = summarize(computeScope(groups, device, {}, { decided: {} }));

		expect(small.na).toBeGreaterThan(big.na);
		expect(big.na).toBe(0);
		expect(small.emptyGroups).toBe(0);
	});
});
