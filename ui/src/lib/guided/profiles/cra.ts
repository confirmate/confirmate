import type { FrameworkProfile } from '../profile';

/**
 * The CRA profile. The texts of the classes and routes are sample content until they are reviewed together with the
 * catalog.
 */
export const craProfile: FrameworkProfile = {
	id: 'cra',
	name: 'Cyber Resilience Act',
	kind: 'PRODUCT',
	kindLabel: 'Product',
	pitch: {
		title: 'A product',
		text: 'Software or firmware that you place on the EU market. We guide you through classification, risk assessment and conformity.'
	},
	steps: [
		{
			id: 'product',
			type: 'details',
			label: 'Product',
			title: 'Tell us about the product',
			intro:
				'The Cyber Resilience Act applies to a product, not to a company. These details identify it and set the support period you have to keep.'
		},
		{
			id: 'classification',
			type: 'choice',
			label: 'Classification',
			title: 'What kind of product is it?',
			intro:
				'The CRA treats some products as more critical than others. This decides how strictly you have to prove conformity. Pick the description that fits the main function of the product best.'
		},
		{
			id: 'risk',
			type: 'questionnaire',
			label: 'Risk assessment',
			title: 'Assess the risks of the product',
			intro:
				'The CRA asks for a cybersecurity risk assessment. It also decides which of the product requirements apply to you. Most of them always do.'
		},
		{
			id: 'scope',
			type: 'scope',
			label: 'Scope',
			title: 'Confirm what applies',
			intro: 'The reason for every requirement that drops out is recorded. You are responsible for this decision, so please read it before you approve.'
		},
		{
			id: 'sources',
			type: 'sources',
			label: 'Connect sources',
			title: 'Show us where the evidence is',
			intro:
				'Confirmate reads from the tools you already use. It never changes anything in them. Skip a source and we add a task for it instead.'
		}
	],
	details: [
		{ id: 'name', label: 'Product name', type: 'text', defaultValue: '' },
		{
			id: 'description',
			label: 'What does it do?',
			hint: 'One sentence. It appears on the declaration of conformity.',
			type: 'text',
			defaultValue: ''
		},
		{ id: 'version', label: 'Current release', type: 'text', defaultValue: '' },
		{
			id: 'kind',
			label: 'Kind of product',
			type: 'select',
			options: [
				'Software application',
				'Library or component',
				'Firmware or embedded software',
				'Online service with a client'
			],
			defaultValue: 'Software application'
		},
		{
			id: 'support',
			label: 'Support until',
			hint: 'At least five years, unless the product is expected to be used for less. We remind you before it ends.',
			type: 'date',
			defaultValue: ''
		}
	],
	choices: [
		{
			id: 'none',
			label: 'A general-purpose product',
			hint: 'Nothing on the list matches',
			result: {
				title: 'Default class',
				text: 'You assess conformity yourself (internal production control) and sign the declaration of conformity.',
				note: 'Sample text. The route for each class is checked with the catalog review.'
			}
		},
		{
			id: 'password-manager',
			label: 'Password manager',
			hint: 'Stores or generates credentials for users',
			result: {
				title: 'Important, class I',
				text: 'Self-assessment is only possible if you apply the relevant standards in full. Otherwise a third-party assessment is needed.',
				note: 'Sample text. Your product is listed in the category of important products, class I.'
			}
		},
		{
			id: 'vpn',
			label: 'VPN or network management software',
			hint: 'Creates tunnels or manages network devices',
			result: {
				title: 'Important, class I',
				text: 'Self-assessment is only possible if you apply the relevant standards in full. Otherwise a third-party assessment is needed.',
				note: 'Sample text. Your product is listed in the category of important products, class I.'
			}
		},
		{
			id: 'hypervisor',
			label: 'Hypervisor or container runtime',
			hint: 'Isolates and runs other software',
			result: {
				title: 'Important, class II',
				text: 'A third-party assessment is required before you can place the product on the market.',
				note: 'Sample text. Your product is listed in the category of important products, class II.'
			}
		}
	],
	questionnaire: {
		heading: 'Product requirements, Annex I, part I',
		questions: [
			{ id: 'data', text: 'Does the product process or store data of users, or data about them?', help: 'Credentials, personal data, usage data', defaultAnswer: 'yes' },
			{ id: 'net', text: 'Does it have a network interface or any other external interface?', help: 'Network, USB, serial, bluetooth, an API', defaultAnswer: 'yes' },
			{ id: 'acc', text: 'Does it have user accounts or a login?', help: 'Any form of authentication', defaultAnswer: 'yes' },
			{ id: 'upd', text: 'Can it be updated after it has been delivered?', help: 'Remotely or locally', defaultAnswer: 'yes' },
			{ id: 'dbg', text: 'Is there a debug, service or manufacturing interface?', help: 'Typical for firmware and devices', defaultAnswer: 'no' }
		],
		requirements: [
			{ id: 'a', title: 'No known exploitable vulnerabilities' },
			{ id: 'b', title: 'Secure by default' },
			{ id: 'c', title: 'Security updates' },
			{ id: 'd', title: 'Protection from unauthorised access' },
			{ id: 'e', title: 'Confidentiality of data' },
			{ id: 'f', title: 'Integrity' },
			{ id: 'g', title: 'Data minimisation' },
			{ id: 'h', title: 'Availability' },
			{ id: 'i', title: 'Impact on other services' },
			{ id: 'j', title: 'Limited attack surface' },
			{ id: 'k', title: 'Exploitation mitigation' },
			{ id: 'l', title: 'Logging and monitoring' },
			{ id: 'm', title: 'Removal and transfer of data' }
		],
		effects: [
			{ when: { data: 'no' }, out: ['e', 'g', 'm'], reason: 'No user data is processed.' },
			{ when: { net: 'no' }, out: ['h', 'i'], reason: 'No network or external interface.' },
			{ when: { acc: 'no', net: 'no' }, out: ['d'], reason: 'No accounts and no external interface.' }
		],
		notes: [
			{
				when: { upd: 'no' },
				tone: 'warn',
				text: 'A product that cannot be updated can rarely meet the requirements on security updates and on vulnerability handling. Think about how you will deliver fixes.'
			},
			{
				when: { dbg: 'yes' },
				tone: 'brand',
				text: 'Requirements for protecting debug and service interfaces will be added to your scope.'
			}
		]
	},
	sources: [
		{
			id: 'code',
			title: 'Source code and builds',
			text: 'GitHub or GitLab. We read protection of the main branch, reviews, signed commits, pipeline runs, releases and the security policy.',
			state: 'connected',
			stateLabel: 'Connected',
			detail: 'Sample: pythonpass/pythonpass, last read 4 minutes ago, read-only access.'
		},
		{
			id: 'components',
			title: 'Components and vulnerabilities',
			text: 'Dependency-Track. We found one project that matches your package: pythonpass 1.4.0.',
			state: 'connected',
			stateLabel: 'Connected',
			actions: ['Use this project', 'Choose another']
		},
		{
			id: 'issues',
			title: 'Security issues',
			text: 'Issues and security advisories of the repository. They show how fast vulnerabilities are handled.',
			state: 'connected',
			stateLabel: 'Connected'
		},
		{
			id: 'documents',
			title: 'Documents',
			text: 'Risk assessment, vulnerability disclosure policy, user information, update distribution, support period and declaration of conformity. Upload what exists, we help you write the rest later.',
			state: 'partial',
			stateLabel: '3 of 6 added',
			actions: ['Upload documents']
		},
		{
			id: 'pipeline',
			title: 'Results from your build pipeline',
			text: 'SBOMs, scan results and repository checks can be sent from your pipeline with one step. It works with any tool that writes the common formats.',
			state: 'optional',
			stateLabel: 'Optional',
			snippet: '- uses: confirmate/send-results@v1\n  with:\n    product: pythonpass\n    files: sbom.cdx.json, results.sarif'
		}
	],
	alwaysApplies: {
		count: 30,
		text: 'The risk assessment itself, handling of vulnerabilities (Annex I, part II), the obligations of manufacturers, reporting of exploited vulnerabilities and incidents, and the technical documentation with the declaration of conformity.'
	}
};
