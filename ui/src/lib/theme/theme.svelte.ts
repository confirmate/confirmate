import { browser } from '$app/environment';

/** The themes that are defined in theme.css. */
export const themes = [
	{ id: 'console', label: 'Console' },
	{ id: 'console-inter', label: 'Console (Inter)' },
	{ id: 'editorial', label: 'Editorial' },
	{ id: 'night', label: 'Night' }
] as const;

export type ThemeId = (typeof themes)[number]['id'];

const storageKey = 'cf.theme';

function load(): ThemeId {
	if (!browser) return 'console';
	try {
		const stored = localStorage.getItem(storageKey);
		if (themes.some((t) => t.id === stored)) return stored as ThemeId;
	} catch {
		// Storage can be unavailable, the default theme is fine then.
	}
	return 'console';
}

class ThemeState {
	current = $state<ThemeId>(load());

	/** Applies the theme to the document and remembers the choice. */
	set(id: ThemeId) {
		this.current = id;
		this.apply();
		try {
			localStorage.setItem(storageKey, id);
		} catch {
			// Not persisting is acceptable.
		}
	}

	/** Writes the current theme to the data-theme attribute of the document. */
	apply() {
		if (browser) document.documentElement.dataset.theme = this.current;
	}
}

export const theme = new ThemeState();
