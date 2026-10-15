import { conditionsOf } from './catalog';
import { suggestedAnswers } from './conditions';
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
	/** The conditions whose answer a person has set, so that a change of the product type leaves them alone. */
	answered = $state<Record<string, boolean>>({});
	/** The conditions whose answer comes from the product type. */
	fromChoice = $state<Record<string, boolean>>({});
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
		this.answered = {};
		this.answers = {};
		this.applyChoices(profile);
		this.approved = false;
		this.decisions = { decided: {} };
	}

	/**
	 * Sets the answers about the conditions that the choices of the setup suggest, for example the kind of product. An
	 * answer that a person has set is left alone. A requirement stays unless the answer is an explicit no, so "yes" is the
	 * cautious start.
	 */
	applyChoices(profile: FrameworkProfile) {
		const suggested = suggestedAnswers(profile, this.details, conditionsOf(this.groups));
		for (const [id, s] of Object.entries(suggested)) {
			if (this.answered[id]) continue;
			this.answers[id] = s.answer;
			this.fromChoice[id] = s.fromChoice;
		}
		this.approved = false;
	}

	/** Records the answer of a person about a condition. */
	answer(id: string, answer: Answer) {
		this.answers[id] = answer;
		this.answered[id] = true;
		this.fromChoice[id] = false;
		this.approved = false;
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
