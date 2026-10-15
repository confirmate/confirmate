import { browser } from '$app/environment';

/**
 * Feature flags of the UI. Everything new is off by default, so that existing deployments keep the current screens.
 *
 * A flag is on if it is set at build time (VITE_FEATURE_<NAME>=true) or, for trying it out, in the browser: open any
 * page with ?<name>=on (or ?<name>=off to switch it off again), which is remembered in the browser, or set
 * localStorage.setItem('cf.feature.<name>', 'true').
 */
function enabled(name: string, buildTime: unknown): boolean {
	if (buildTime === 'true') return true;
	if (!browser) return false;
	try {
		const requested = new URLSearchParams(location.search).get(name);
		if (requested === 'on' || requested === 'off') {
			localStorage.setItem(`cf.feature.${name}`, requested === 'on' ? 'true' : 'false');
		}
		return localStorage.getItem(`cf.feature.${name}`) === 'true';
	} catch {
		return false;
	}
}

/** The guided experience (targets of evaluation, setup wizard, requirement view). */
export const guidedEnabled = () => enabled('guided', import.meta.env.VITE_FEATURE_GUIDED);
