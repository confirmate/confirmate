import { evidenceClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
  const evClient = evidenceClient(fetch);

  const evidenceRes = await evClient.GET('/v1/evidence_store/evidences', {
    params: {
      query: {
        'filter.targetOfEvaluationId': params.id,
        pageSize: 1000
      }
    }
  });

  if (!evidenceRes.response.ok) {
    const text = await evidenceRes.response.text();
    console.error('Evidence API error:', evidenceRes.response.status, text);
    error(evidenceRes.response.status, 'Failed to load evidences');
  }

  return {
    evidences: evidenceRes.data?.evidences ?? []
  };
};