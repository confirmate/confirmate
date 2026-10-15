import type { RequirementGroup } from '../profile';

/**
 * The product requirements of the CRA (Annex I, part I) as the setup wizard offers them for scoping. They come from the CRA
 * catalog in example-data (the control short names are the IDs). The applicability (mandatory, only if a capability exists,
 * risk based) is a sample classification until it is reviewed together with the catalog.
 */
export const craRequirementGroups: RequirementGroup[] = [
	{
		id: "CRA-I-1.1.1",
		title: "Appropriate level of cybersecurity based on the risks",
		reference: "Annex I, part I, (1)",
		requirements: [
			{
				id: "CRA-I-1.1.1.1",
				title: "Cybersecurity risk assessment",
				text: "A cybersecurity risk assessment of the product exists and is taken into account in planning, design, development, production, delivery and maintenance.",
				checks: 1,
				applicability: "mandatory"
			},
			{
				id: "CRA-I-1.1.1.2",
				title: "Risk assessment is kept up to date",
				text: "The risk assessment is reviewed and updated during the support period, in particular after relevant vulnerabilities and changes.",
				checks: 1,
				applicability: "mandatory"
			},
		]
	},
	{
		id: "CRA-I-1.1.2",
		title: "No Known Exploitable Vulnerabilities",
		reference: "Annex I, part I, (2)(a)",
		requirements: [
			{
				id: "CRA-I-1.1.2.1",
				title: "Vulnerability Scanning",
				text: "Regular vulnerability scanning must be performed",
				checks: 3,
				applicability: "mandatory"
			},
			{
				id: "CRA-I-1.1.2.2",
				title: "Known vulnerabilities by severity",
				text: "The number of known vulnerabilities of the product stays within the accepted limits per severity.",
				checks: 4,
				applicability: "mandatory"
			},
		]
	},
	{
		id: "CRA-I-1.1.3",
		title: "Secure by Default Configuration",
		reference: "Annex I, part I, (2)(b)",
		requirements: [
			{
				id: "CRA-I-1.1.3.1",
				title: "Secure Default Configuration",
				text: "Secure configuration must be enforced by default",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.3.2",
				title: "Reset to original state",
				text: "The product can be reset to its original, secure state.",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.3.3",
				title: "No default or shared credentials",
				text: "The product does not ship with default or shared credentials; credentials are unique per device or are set at first use.",
				checks: 1,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.4",
		title: "Security Updates",
		reference: "Annex I, part I, (2)(c)",
		requirements: [
			{
				id: "CRA-I-1.1.4.1",
				title: "Update Check Interval",
				text: "Automatic updates must check for updates within appropriate intervals",
				checks: 1,
				applicability: "capability",
				capability: { question: "upd", text: "the product can be updated" }
			},
			{
				id: "CRA-I-1.1.4.2",
				title: "Automatic Updates by Default",
				text: "Automatic updates must be enabled by default",
				checks: 1,
				applicability: "capability",
				capability: { question: "upd", text: "the product can be updated" }
			},
			{
				id: "CRA-I-1.1.4.3",
				title: "User Update Control",
				text: "Users must be able to postpone updates and receive notifications",
				checks: 1,
				applicability: "capability",
				capability: { question: "upd", text: "the product can be updated" }
			},
		]
	},
	{
		id: "CRA-I-1.1.5",
		title: "Access Control and Unauthorized Access Protection",
		reference: "Annex I, part I, (2)(d)",
		requirements: [
			{
				id: "CRA-I-1.1.5.1",
				title: "Access Control Mechanisms",
				text: "Access control mechanisms must be implemented at critical points",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.5.2",
				title: "Password Policies",
				text: "Password policies must be enabled for identity management",
				checks: 1,
				applicability: "capability",
				capability: { question: "acc", text: "the product has user accounts" }
			},
			{
				id: "CRA-I-1.1.5.3",
				title: "Unauthorized Access Reporting",
				text: "Unauthorized access attempts must be reported",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.5.4",
				title: "Multi-Factor Authentication",
				text: "Multi-factor authentication must be enabled for administrative access",
				checks: 1,
				applicability: "capability",
				capability: { question: "acc", text: "the product has user accounts" }
			},
			{
				id: "CRA-I-1.1.5.5",
				title: "Anomaly Detection",
				text: "Anomaly detection mechanisms must be enabled",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.5.6",
				title: "Debug and service interfaces are protected",
				text: "Debug, service and manufacturing interfaces are disabled, locked or protected by authentication in the delivered product.",
				checks: 1,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.6",
		title: "Data Confidentiality Protection",
		reference: "Annex I, part I, (2)(e)",
		requirements: [
			{
				id: "CRA-I-1.1.6.1",
				title: "Encryption at Rest",
				text: "Data at rest must be encrypted",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.6.2",
				title: "Cryptographic Algorithm Strength",
				text: "Strong cryptographic algorithms must be used",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.6.3",
				title: "Encryption in Transit",
				text: "Data in transit must be encrypted",
				checks: 1,
				applicability: "capability",
				capability: { question: "net", text: "the product sends data over a network" }
			},
			{
				id: "CRA-I-1.1.6.4",
				title: "Cryptographic keys are stored protected",
				text: "Keys and secrets are stored in protected storage (for example secure element, key store) and not in clear text in the product or its firmware image.",
				checks: 1,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.7",
		title: "Data Integrity Protection",
		reference: "Annex I, part I, (2)(f)",
		requirements: [
			{
				id: "CRA-I-1.1.7.1",
				title: "Secure Transmission Protocols",
				text: "Secure protocols must be enabled for data transmission",
				checks: 1,
				applicability: "capability",
				capability: { question: "net", text: "the product sends data over a network" }
			},
			{
				id: "CRA-I-1.1.7.2",
				title: "Password Policy Enforcement",
				text: "Password policies must be enforced",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.7.3",
				title: "Integrity of software and firmware",
				text: "Software and firmware are protected against unauthorized modification, for example by signed updates, secure boot and verified update packages.",
				checks: 2,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.7.4",
				title: "Source integrity",
				text: "Changes to the source code are traceable and protected: commits are signed and changes are reviewed.",
				checks: 3,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.8",
		title: "Data Minimization",
		reference: "Annex I, part I, (2)(g)",
		requirements: [
			{
				id: "CRA-I-1.1.8.1",
				title: "Data Minimization",
				text: "Data minimization techniques must be enabled",
				checks: 1,
				applicability: "mandatory"
			},
			{
				id: "CRA-I-1.1.8.2",
				title: "Automated Data Deletion",
				text: "Automated deletion of data must be configured",
				checks: 1,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.9",
		title: "Availability and DoS Protection",
		reference: "Annex I, part I, (2)(h)",
		requirements: [
			{
				id: "CRA-I-1.1.9.1",
				title: "DDoS Protection",
				text: "DDoS protection mechanisms must be activated",
				checks: 1,
				applicability: "capability",
				capability: { question: "net", text: "the product offers a network service" }
			},
			{
				id: "CRA-I-1.1.9.2",
				title: "Essential functions stay available",
				text: "Essential and basic functions remain available after an incident, including resilience against denial of service.",
				checks: 1,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.10",
		title: "Minimize Impact on Other Services",
		reference: "Annex I, part I, (2)(i)",
		requirements: [
			{
				id: "CRA-I-1.1.10.1",
				title: "Interface and Protocol Inventory",
				text: "Standard protocols must be listed and only required interfaces enabled",
				checks: 1,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.11",
		title: "Limit Attack Surfaces",
		reference: "Annex I, part I, (2)(j)",
		requirements: [
			{
				id: "CRA-I-1.1.11.1",
				title: "Risk Assessment Mapping",
				text: "Risk assessment must be mapped to services and controls",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.11.2",
				title: "Minimal Attack Surface",
				text: "Only required interfaces and protocols must be enabled",
				checks: 1,
				applicability: "mandatory"
			},
		]
	},
	{
		id: "CRA-I-1.1.12",
		title: "Exploitation Mitigation",
		reference: "Annex I, part I, (2)(k)",
		requirements: [
			{
				id: "CRA-I-1.1.12.1",
				title: "Critical Data Backup",
				text: "Backup mechanisms must be enabled for critical data flows",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.12.2",
				title: "Exploitation mitigations are applied",
				text: "Hardening and exploitation mitigations are applied, for example memory protection, compiler hardening flags, sandboxing and least privilege.",
				checks: 1,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.13",
		title: "Security Logging and Monitoring",
		reference: "Annex I, part I, (2)(l)",
		requirements: [
			{
				id: "CRA-I-1.1.13.1",
				title: "Security Activity Logging",
				text: "Activity logging must be enabled for security-relevant activities",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.13.2",
				title: "Identity Activity Tracking",
				text: "Identity activity must be tracked and available",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.13.3",
				title: "Logging Opt-Out",
				text: "Logging opt-out mechanism must be available to users",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.13.4",
				title: "Security Data Access",
				text: "Security data must be available to users",
				checks: 1,
				applicability: "risk"
			},
		]
	},
	{
		id: "CRA-I-1.1.14",
		title: "Secure Data Deletion",
		reference: "Annex I, part I, (2)(m)",
		requirements: [
			{
				id: "CRA-I-1.1.14.1",
				title: "Secure Data Deletion",
				text: "Secure data deletion mechanism must be activated",
				checks: 1,
				applicability: "risk"
			},
			{
				id: "CRA-I-1.1.14.2",
				title: "Secure transfer of data",
				text: "Where data can be transferred to other products or systems, the transfer is secure.",
				checks: 1,
				applicability: "risk"
			},
		]
	},
];
