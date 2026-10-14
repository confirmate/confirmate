import { browser } from '$app/environment';

export type SourceMode = 'real' | 'sample';

const storageKey = 'cf.guided.source';

function load(): SourceMode {
	if (!browser) return 'real';
	try {
		return localStorage.getItem(storageKey) === 'sample' ? 'sample' : 'real';
	} catch {
		return 'real';
	}
}

/** Whether the guided screens show the data of the orchestrator or sample data. */
class SourceModeState {
	current = $state<SourceMode>(load());

	set(mode: SourceMode) {
		this.current = mode;
		try {
			localStorage.setItem(storageKey, mode);
		} catch {
			// Not persisting is acceptable.
		}
	}
}

export const sourceMode = new SourceModeState();
