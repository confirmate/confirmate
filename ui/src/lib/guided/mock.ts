import type { GuidedSource, TargetSummary } from './source';

/** Sample data for the preview of the guided experience. It is replaced by the real source. */
const targets: TargetSummary[] = [
	{
		id: 'pythonpass',
		name: 'Pythonpass',
		subtitle: 'Password manager, release 1.4.0',
		kindLabel: 'Product',
		frameworkLabel: 'CRA, important class I',
		tags: ['GitHub', 'Dependency-Track', 'Python'],
		progress: { met: 32, attention: 8, open: 24, total: 64 },
		nextDeadline: 'Early warning for an actively exploited vulnerability, due in 19 hours'
	},
	{
		id: 'beacon-edge',
		name: 'Beacon Edge',
		subtitle: 'Gateway firmware, release 2.1.3',
		kindLabel: 'Product',
		frameworkLabel: 'CRA, default class',
		tags: ['GitLab', 'Jira', 'C, Yocto'],
		progress: { met: 21, attention: 14, open: 19, total: 54 },
		nextDeadline: 'Review of the risk assessment, in 41 days'
	},
	{
		id: 'acme-cloud',
		name: 'Acme Cloud Platform',
		subtitle: 'Cloud service',
		kindLabel: 'Cloud service',
		frameworkLabel: 'BSI C5',
		tags: [],
		progress: { met: 71, attention: 15, open: 26, total: 112 }
	},
	{
		id: 'acme-devices',
		name: 'Acme Devices GmbH',
		subtitle: 'Organisation',
		kindLabel: 'Organisation',
		frameworkLabel: 'NIS-2',
		tags: [],
		progress: { met: 18, attention: 8, open: 14, total: 40 }
	}
];

export const mockSource: GuidedSource = {
	async listTargets() {
		return targets;
	}
};
