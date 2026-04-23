import { evidenceClient } from '$lib/api/client';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params, fetch }) => {
  const evClient = evidenceClient(fetch);

  const evidenceRes = await evClient.GET('/v1/evidence_store/evidences', {
    params: {
      query: {
        targetOfEvaluationId: params.id,
        pageSize: 1000
      }
    }
  });

  return {
    evidences: evidenceRes.data?.evidences ?? []
  };
};