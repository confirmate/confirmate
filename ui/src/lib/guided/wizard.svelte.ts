import type { Answer, FrameworkProfile } from './profile';

/** The answers of the setup wizard. They live across the steps and are sent to the backend at the end. */
class WizardState {
	details = $state<Record<string, string>>({});
	choiceId = $state('');
	answers = $state<Record<string, Answer>>({});
	approved = $state(false);

	/** Resets the state to the defaults of the profile. */
	start(profile: FrameworkProfile) {
		this.details = Object.fromEntries(profile.details.map((d) => [d.id, d.defaultValue]));
		this.choiceId = profile.choices[0]?.id ?? '';
		this.answers = Object.fromEntries(profile.questionnaire.questions.map((q) => [q.id, q.defaultAnswer]));
		this.approved = false;
	}

	/** True once the wizard has been started for a profile. */
	get started() {
		return Object.keys(this.details).length > 0;
	}
}

export const wizard = new WizardState();
