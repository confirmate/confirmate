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

export function findResource<K extends ResourceType>(
    evidences: SchemaEvidence[],
    type: K
): { evidence: SchemaEvidence; data: NonNullable<SchemaResource[K]> } | null {
    const evidence = evidences.find((e) => getResourceType(e) === type);
    if (!evidence) {
        return null;
    }
    const data = (evidence.resource as SchemaResource)[type];
    if (!data) {
        return null;
    }
    return {evidence, data: data as NonNullable<SchemaResource[K]>};
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