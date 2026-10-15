import type { FrameworkProfile } from '../profile';

/**
 * The CRA profile. The texts of the classes and routes are sample content until they are reviewed together with the
 * catalog.
 */
export const craProfile: FrameworkProfile = {
	id: 'cra',
	name: 'Cyber Resilience Act',
	catalogId: 'CRA_Catalog',
	targetType: 'TARGET_TYPE_PRODUCT',
	typeLabel: 'Product',
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
			type: 'risk',
			label: 'Risk assessment',
			title: 'Assess the risks of the product',
			intro:
				'The CRA asks for a cybersecurity risk assessment, and it decides which requirements apply to your product. Say what the product protects and where it runs. The requirements that depend on the risk follow from that.'
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
			defaultValue: 'Software application',
			// What the kind of product says about the conditions of the requirements. These are suggestions for the risk step.
			conditionDefaults: {
				'Software application': {
					'The product has debug, service or manufacturing interfaces': 'no'
				},
				'Library or component': {
					'Automatic security updates are appropriate for the product': 'no',
					'Users or administrators authenticate with passwords': 'no',
					'The product has debug, service or manufacturing interfaces': 'no',
					'The product uses cryptography': 'unknown',
					'The product holds cryptographic keys or secrets': 'unknown',
					'The product processes personal or other user data': 'unknown',
					'The product connects to other devices or networks': 'no',
					'The product has user accounts': 'no',
					'The product records security relevant activity': 'no',
					'The product stores user data or settings': 'no',
					'Data can be transferred to other products or systems': 'no'
				},
				'Firmware or embedded software': {
					'The product has user accounts': 'unknown',
					'The product processes personal or other user data': 'unknown'
				},
				'Online service with a client': {
					'The product has debug, service or manufacturing interfaces': 'no',
					'Separating security updates from functionality updates is technically feasible': 'unknown'
				}
			}
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
			attributes: { classification: 'CLASSIFICATION_DEFAULT' },
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
			attributes: { classification: 'CLASSIFICATION_IMPORTANT_I' },
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
			attributes: { classification: 'CLASSIFICATION_IMPORTANT_I' },
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
			attributes: { classification: 'CLASSIFICATION_IMPORTANT_II' },
			label: 'Hypervisor or container runtime',
			hint: 'Isolates and runs other software',
			result: {
				title: 'Important, class II',
				text: 'A third-party assessment is required before you can place the product on the market.',
				note: 'Sample text. Your product is listed in the category of important products, class II.'
			}
		}
	],
	risk: {
		heading: 'Requirements of the CRA',
		assetSuggestions: ['User credentials', 'User data', 'Firmware or program code', 'Configuration', 'Keys and secrets'],
		presets: [
			{
				id: 'connected',
				label: 'A connected app or service',
				hint: 'Used by consumers or customers over the internet',
				profile: {
					assets: [
						{ name: 'Account credentials', confidentiality: 4, integrity: 4, availability: 2 },
						{ name: 'User content', confidentiality: 3, integrity: 3, availability: 3 }
					],
					physicalAccess: 'SUPERVISED',
					connectivity: 'INTERNET',
					operatorSkill: 'GENERAL'
				}
			},
			{
				id: 'desktop',
				label: 'A desktop or command-line tool',
				hint: 'Runs on the computer of the user, mostly offline',
				profile: {
					assets: [{ name: 'User data', confidentiality: 3, integrity: 2, availability: 1 }],
					physicalAccess: 'CONTROLLED',
					connectivity: 'LOCAL',
					operatorSkill: 'GENERAL'
				}
			},
			{
				id: 'business',
				label: 'Business software in a company network',
				hint: 'Operated by administrators in a private network',
				profile: {
					assets: [{ name: 'Business data', confidentiality: 4, integrity: 4, availability: 3 }],
					physicalAccess: 'CONTROLLED',
					connectivity: 'PRIVATE_NETWORK',
					operatorSkill: 'EXPERT'
				}
			},
			{
				id: 'device',
				label: 'A device or firmware in the field',
				hint: 'Installed where anybody can reach it, connected to the internet',
				profile: {
					assets: [
						{ name: 'Device credentials and keys', confidentiality: 5, integrity: 5, availability: 3 },
						{ name: 'Firmware', confidentiality: 2, integrity: 5, availability: 4 }
					],
					physicalAccess: 'OPEN',
					connectivity: 'INTERNET',
					operatorSkill: 'GENERAL'
				}
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
};
