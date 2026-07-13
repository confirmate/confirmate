import type {SchemaEvidence, SchemaResource} from '$lib/api/openapi/evidence';

export type ResourceType = keyof SchemaResource;

export function getResourceType(evidence: SchemaEvidence): ResourceType | 'unknown' {
    const resource = evidence.resource;
    if (!resource) {
        return 'unknown';
    }
    const key = Object.keys(resource)[0] as ResourceType | undefined;
    return key ?? 'unknown';
}

// findResource accepts any string key so the document templates can still
// reference resource types that come and go in the security-metrics ontology
// without breaking the build. Unknown types return null; the templates handle
// the fallback path via optional chaining.
//
// The returned `data` is typed as `Record<string, any>` so accessing fields
// that may have been renamed/removed in the ontology submodule doesn't error
// at type-check time — runtime access still degrades cleanly to undefined.
export function findResource(
    evidences: SchemaEvidence[],
    type: string
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
): { evidence: SchemaEvidence; data: Record<string, any> } | null {
    const evidence = evidences.find((e) => getResourceType(e) === type);
    if (!evidence || !evidence.resource) {
        return null;
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const data = (evidence.resource as Record<string, any>)[type];
    if (!data) {
        return null;
    }
    return { evidence, data };
}

export interface ManufacturerInfo {
    name: string;
    tradeName: string;
    postalAddress: string;
    generalEmail: string;
    securityEmail: string;
    website: string;
    securityPortalUrl: string;
}

export interface Placeholders {
    userInstructionsReference: string;
    automaticUpdateMethod: string;
    dataRemovalMethod: string;
    disableUpdatesPath: string;
    integrationDocUrl: string;
    architectureDocumentReference: string;
    hardwareLayoutReference: string;
    updateInstructionsUrl: string;
}