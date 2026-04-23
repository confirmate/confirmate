import type { SchemaEvidence as Evidence } from '$lib/api/openapi/evidence';

interface PopoverState {
  evidence: Evidence | null;
  resourceType: string;
  field: string;
  anchor: HTMLElement | null;
}

export const popover = $state<PopoverState>({
  evidence: null,
  resourceType: '',
  field: '',
  anchor: null
});

export function openPopover(
  evidence: Evidence,
  resourceType: string,
  field: string,
  anchor: HTMLElement
) {
  popover.evidence = evidence;
  popover.resourceType = resourceType;
  popover.field = field;
  popover.anchor = anchor;
}

export function closePopover() {
  popover.evidence = null;
  popover.anchor = null;
}

export function activeKey(): string | null {
  return popover.evidence ? `${popover.resourceType}.${popover.field}` : null;
}