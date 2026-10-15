import type { Answer, FrameworkProfile } from './profile';
import type { Decisions, Treatment } from './scope';

/** The answers of the setup wizard. They live across the steps and are sent to the backend at the end. */
class WizardState {
	details = $state<Record<string, string>>({});
	choiceId = $state('');
	answers = $state<Record<string, Answer>>({});
	approved = $state(false);
	/** The decisions of the person about single requirements, and the suggestions that were dismissed. */
	decisions = $state<Decisions>({ decided: {}, dismissed: {} });

	/** Resets the state to the defaults of the profile. */
	start(profile: FrameworkProfile) {
		this.details = Object.fromEntries(profile.details.map((d) => [d.id, d.defaultValue]));
		this.choiceId = profile.choices[0]?.id ?? '';
		this.answers = Object.fromEntries(profile.questionnaire.questions.map((q) => [q.id, q.defaultAnswer]));
		this.approved = false;
		this.decisions = { decided: {}, dismissed: {} };
	}

	/** Records what a person decided for a requirement. */
	decide(id: string, treatment: Treatment, reason: string) {
		this.decisions.decided[id] = { treatment, reason };
		this.approved = false;
	}

	/** Changes the reason of a decision that was made before. */
	setReason(id: string, reason: string) {
		const d = this.decisions.decided[id];
		if (d) d.reason = reason;
	}

	/** Keeps a requirement as it is and drops the suggestion for it. */
	dismiss(id: string) {
		this.decisions.dismissed[id] = true;
		this.approved = false;
	}

	/** True once the wizard has been started for a profile. */
	get started() {
		return Object.keys(this.details).length > 0;
	}
}

export const wizard = new WizardState();
