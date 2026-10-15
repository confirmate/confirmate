import { conditionsOf } from './catalog';
import type { Answer, FrameworkProfile, RequirementGroup, RiskProfile } from './profile';
import { emptyRisk } from './risk';
import type { Decisions, NotApplicableReason, Treatment } from './scope';

function copyRisk(r: RiskProfile): RiskProfile {
	return { ...r, assets: r.assets.map((a) => ({ ...a })) };
}

/** The answers of the setup wizard. They live across the steps and are sent to the backend at the end. */
class WizardState {
	details = $state<Record<string, string>>({});
	choiceId = $state('');
	/** The requirements of the catalog, as the scope is worked out for them. */
	groups = $state.raw<RequirementGroup[]>([]);
	/** The assets and the environment of the risk assessment. */
	risk = $state<RiskProfile>(emptyRisk());
	presetId = $state('');
	/** The answers about the conditions of the requirements, by condition ID. */
	answers = $state<Record<string, Answer>>({});
	approved = $state(false);
	/** The decisions of the person about single requirements. */
	decisions = $state<Decisions>({ decided: {} });

	/** Resets the state to the defaults of the profile. */
	start(profile: FrameworkProfile, groups: RequirementGroup[]) {
		const preset = profile.risk.presets[0];

		this.details = Object.fromEntries(profile.details.map((d) => [d.id, d.defaultValue]));
		this.choiceId = profile.choices[0]?.id ?? '';
		this.groups = groups;
		this.risk = preset ? copyRisk(preset.profile) : emptyRisk();
		this.presetId = preset?.id ?? '';
		// A requirement stays if the answer is not an explicit no, so "yes" is the cautious start.
		this.answers = Object.fromEntries(conditionsOf(groups).map((c) => [c.id, 'yes' as Answer]));
		this.approved = false;
		this.decisions = { decided: {} };
	}

	/** Takes over a ready-made risk profile, which the person can adjust afterwards. */
	usePreset(profile: FrameworkProfile, id: string) {
		const preset = profile.risk.presets.find((p) => p.id === id);
		if (!preset) return;
		this.presetId = id;
		this.risk = copyRisk(preset.profile);
		this.approved = false;
	}

	/** Records what a person decided for a requirement. */
	decide(id: string, treatment: Treatment, reason: string, naReason?: NotApplicableReason, compensatedBy?: string) {
		this.decisions.decided[id] = { treatment, reason, naReason, compensatedBy };
		this.approved = false;
	}

	/** Changes the reason of a decision that was made before. */
	setReason(id: string, reason: string) {
		const d = this.decisions.decided[id];
		if (d) d.reason = reason;
	}

	/** Forgets the decision about a requirement, so that the rules apply again. */
	reset(id: string) {
		delete this.decisions.decided[id];
		this.approved = false;
	}

	/** True once the wizard has been started for a profile. */
	get started() {
		return Object.keys(this.details).length > 0;
	}
}

export const wizard = new WizardState();
