import type { GuidedSource, TargetDetail, TargetSummary } from './source';

/** Sample data for the preview of the guided experience. It is replaced by the real source. */
const targets: TargetSummary[] = [
	{
		id: 'pythonpass',
		name: 'Pythonpass',
		subtitle: 'Password manager, release 1.4.0',
		typeLabel: 'Product',
		frameworkLabel: 'CRA, important class I',
		tags: ['GitHub', 'Dependency-Track', 'Python'],
		progress: { met: 32, attention: 8, open: 24, total: 64 },
		nextDeadline: 'Early warning for an actively exploited vulnerability, due in 19 hours'
	},
	{
		id: 'beacon-edge',
		name: 'Beacon Edge',
		subtitle: 'Gateway firmware, release 2.1.3',
		typeLabel: 'Product',
		frameworkLabel: 'CRA, default class',
		tags: ['GitLab', 'Jira', 'C, Yocto'],
		progress: { met: 21, attention: 14, open: 19, total: 54 },
		nextDeadline: 'Review of the risk assessment, in 41 days'
	},
	{
		id: 'acme-cloud',
		name: 'Acme Cloud Platform',
		subtitle: 'Cloud service',
		typeLabel: 'Cloud service',
		frameworkLabel: 'BSI C5',
		tags: [],
		progress: { met: 71, attention: 15, open: 26, total: 112 }
	},
	{
		id: 'acme-devices',
		name: 'Acme Devices GmbH',
		subtitle: 'Organisation',
		typeLabel: 'Organisation',
		frameworkLabel: 'NIS-2',
		tags: [],
		progress: { met: 18, attention: 8, open: 14, total: 40 }
	}
];

const pythonpass: TargetDetail = {
	...targets[0],
	chips: ['Release 1.4.0', 'Supported until 1 October 2031'],
	clock: {
		title: 'An actively exploited vulnerability may affect this product',
		text: 'The component cryptography 41.0.2 in release 1.4.0 appears in a list of exploited vulnerabilities. Confirm whether the product is affected. If it is, the CRA requires an early warning within 24 hours.',
		action: 'Start the report',
		steps: [
			{ label: '1 Early warning', due: 'due in 19 hours' },
			{ label: '2 Notification', due: 'within 72 hours' },
			{ label: '3 Final report', due: '14 days after the fix' }
		]
	},
	tasks: [
		{ id: 't1', title: 'Default configuration is not secure', detail: 'Secure by default (b). The code check found a debug mode that is on by default.', urgency: 'urgent', assignee: 'Bob', due: 'due Friday', requirementId: 'secure-by-default' },
		{ id: 't2', title: 'Protect the main branch', detail: 'Source integrity. Branch protection is off, reviews are not required.', urgency: 'attention', assignee: 'Bob' },
		{ id: 't3', title: 'The SBOM of release 1.4.0 is incomplete', detail: 'Vulnerability handling (1). Required fields are missing for 12 components.', urgency: 'attention', assignee: 'Bob' },
		{ id: 't4', title: 'Publish a vulnerability disclosure policy', detail: 'Vulnerability handling (5). We found no SECURITY.md. Write it from a template.', urgency: 'open', assignee: 'Alice' },
		{ id: 't5', title: 'Add the user information and instructions', detail: 'Obligations of manufacturers. Upload the document or write it in Confirmate.', urgency: 'open', assignee: 'Alice' }
	],
	parts: [
		{ id: 'product', label: 'Product requirements', met: 18, attention: 5, open: 13, total: 36 },
		{ id: 'vulnerability', label: 'Vulnerability handling', met: 7, attention: 3, open: 3, total: 13 },
		{ id: 'obligations', label: 'Obligations of manufacturers', met: 3, attention: 0, open: 2, total: 5 },
		{ id: 'reporting', label: 'Reporting', met: 1, attention: 0, open: 3, total: 4 },
		{ id: 'documentation', label: 'Documentation and conformity', met: 3, attention: 0, open: 3, total: 6 }
	],
	sources: [
		{ name: 'GitHub', state: 'ok', detail: 'Read 4 minutes ago' },
		{ name: 'Dependency-Track', state: 'ok', detail: 'Read 12 minutes ago' },
		{ name: 'Documents', state: 'partial', detail: '3 of 6 added' },
		{ name: 'Code analysis', state: 'none', detail: 'Not connected' }
	],
	requirements: [
		{ id: 'known-vulnerabilities', partId: 'product', title: 'No known exploitable vulnerabilities', reference: 'Annex I, part I, (2)(a)', status: 'attention' },
		{ id: 'secure-by-default', partId: 'product', title: 'Secure by default', reference: 'Annex I, part I, (2)(b)', status: 'attention' },
		{ id: 'security-updates', partId: 'product', title: 'Security updates', reference: 'Annex I, part I, (2)(c)', status: 'met' },
		{ id: 'access-protection', partId: 'product', title: 'Protection from unauthorised access', reference: 'Annex I, part I, (2)(d)', status: 'met' },
		{ id: 'confidentiality', partId: 'product', title: 'Confidentiality of data', reference: 'Annex I, part I, (2)(e)', status: 'met' },
		{ id: 'integrity', partId: 'product', title: 'Integrity of data, commands and programs', reference: 'Annex I, part I, (2)(f)', status: 'attention' },
		{ id: 'sbom', partId: 'vulnerability', title: 'Software bill of materials', reference: 'Annex I, part II, (1)', status: 'attention' },
		{ id: 'remediation', partId: 'vulnerability', title: 'Remediate vulnerabilities without delay', reference: 'Annex I, part II, (2)', status: 'met' },
		{ id: 'disclosure-policy', partId: 'vulnerability', title: 'Coordinated vulnerability disclosure policy', reference: 'Annex I, part II, (5)', status: 'open' },
		{ id: 'user-information', partId: 'obligations', title: 'User information and instructions', reference: 'Art. 13, Annex II', status: 'open' },
		{ id: 'reporting-process', partId: 'reporting', title: 'Reporting process', reference: 'Art. 14', status: 'open' },
		{ id: 'declaration', partId: 'documentation', title: 'EU declaration of conformity', reference: 'Annex V', status: 'open' }
	],
	requirementDetails: [
		{
			id: 'secure-by-default',
			partId: 'product',
			title: 'Secure by default',
			reference: 'Annex I, part I, (2)(b)',
			status: 'attention',
			plain: 'The product is delivered with a secure default configuration, and users can reset it to that state.',
			scopeReason: 'Applies because of risk assessment, version 1',
			checks: [
				{
					id: 'c1',
					title: 'The default configuration is secure',
					kind: 'automatic',
					status: 'not-met',
					text: 'The code analysis found that debug mode is on in the default settings.',
					code: 'config/default.toml, line 12: debug = true',
					note: 'An automatic result always wins over a manual statement. To change this status, change the code or explain why the finding is wrong.',
					actions: [{ label: 'Open in GitHub' }, { label: 'Report a false finding' }]
				},
				{
					id: 'c2',
					title: 'The product can be reset to its original state',
					kind: 'statement',
					status: 'met',
					text: 'Bob stated on 2 October that this is possible. Evidence: administrator guide, section 4.2.',
					note: 'Valid until 2 October 2027, then we ask again.',
					actions: [{ label: 'Update the statement' }]
				},
				{
					id: 'c3',
					title: 'There are no default or shared credentials',
					kind: 'statement',
					status: 'open',
					text: 'Nobody has answered this yet. You can answer it with a short statement and, if you have one, a document as evidence.',
					actions: [{ label: 'Add a statement', primary: true }, { label: 'Assign to someone' }]
				}
			],
			history: [
				{ when: 'Today', what: 'Code analysis found debug mode on by default' },
				{ when: '2 October', what: 'Bob stated the product can be reset' },
				{ when: '1 October', what: 'Added to the scope by the risk assessment, version 1' }
			],
			comments: [{ who: 'Bob', text: 'The debug flag is needed for the beta build. I will make it a build option.' }]
		}
	]
};

export const mockSource: GuidedSource = {
	async listTargets() {
		return targets;
	},
	async getTarget(id) {
		if (id === pythonpass.id) return pythonpass;
		// The other sample targets have no details in the preview.
		return undefined;
	}
};
