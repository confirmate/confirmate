import { mockSource } from './mock';

/** What the list of targets of evaluation shows about one target. */
export interface TargetSummary {
	id: string;
	name: string;
	subtitle: string;
	kindLabel: string;
	frameworkLabel: string;
	tags: string[];
	progress: { met: number; attention: number; open: number; total: number };
	nextDeadline?: string;
}

/**
 * Where the guided experience gets its data from. Today this is a mock; the real source reads targets of evaluation,
 * evaluation results and deadlines from the orchestrator. The screens only depend on this interface.
 */
export interface GuidedSource {
	listTargets(): Promise<TargetSummary[]>;
}

export const guidedSource: GuidedSource = mockSource;
