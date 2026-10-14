import type { FrameworkProfile } from '../profile';
import { craProfile } from './cra';

/** The profiles that have a guided setup. */
export const profiles: FrameworkProfile[] = [craProfile];

/** Frameworks that are planned. They are shown as not available yet. */
export const plannedProfiles = [
	{
		id: 'c5',
		name: 'BSI C5',
		title: 'A cloud service',
		text: 'A service that you provide to customers. We connect your cloud accounts and build the evidence for an attestation.'
	},
	{
		id: 'nis2',
		name: 'NIS-2',
		title: 'An organisation',
		text: 'A company with obligations for risk management and reporting. We start from the entity type and its sector.'
	}
];
