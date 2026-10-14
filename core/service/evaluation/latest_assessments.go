// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package evaluation

import (
	"context"
	"slices"
	"strings"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
)

// latestAssessments holds the latest assessment result per (resource, metric) pair of one target of evaluation,
// indexed by metric ID. It is fetched once per evaluation of an audit scope, so that the individual (sub-)controls
// do not have to ask the orchestrator over and over again for (slices of) the same data.
//
// The "latest per (resource, metric)" of a pair does not depend on which other pairs are fetched, so selecting the
// metrics of a control from this snapshot yields the same results as filtering by these metrics in the query.
type latestAssessments struct {
	byMetric map[string][]*assessment.AssessmentResult
}

// fetchLatestAssessments fetches the latest assessment result per (resource, metric) pair of the given target of
// evaluation from the orchestrator.
func (svc *Service) fetchLatestAssessments(ctx context.Context, targetOfEvaluationId string) (latest *latestAssessments, err error) {
	var (
		results []*assessment.AssessmentResult
	)

	results, err = api.ListAllPaginated(ctx, &orchestrator.ListAssessmentResultsRequest{
		Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
			TargetOfEvaluationId: &targetOfEvaluationId,
		},
		LatestByResourceId: new(true),
		// Ask for the largest page possible: every additional page makes the orchestrator walk further into the
		// result set again (offset pagination), so fewer, larger pages are cheaper than many small ones.
		PageSize: service.DefaultPaginationOpts.MaxPageSize,
	}, func(ctx context.Context, req *orchestrator.ListAssessmentResultsRequest) (msg *orchestrator.ListAssessmentResultsResponse, err error) {
		var (
			res *connect.Response[orchestrator.ListAssessmentResultsResponse]
		)

		res, err = svc.orchestratorClient.ListAssessmentResults(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}

		return res.Msg, nil
	}, func(res *orchestrator.ListAssessmentResultsResponse) (results []*assessment.AssessmentResult) {
		return res.Results
	})
	if err != nil {
		return nil, err
	}

	return newLatestAssessments(results), nil
}

// newLatestAssessments indexes the given (already latest-per-pair) assessment results by their metric ID.
func newLatestAssessments(results []*assessment.AssessmentResult) (latest *latestAssessments) {
	latest = &latestAssessments{
		byMetric: make(map[string][]*assessment.AssessmentResult),
	}

	for _, r := range results {
		latest.byMetric[r.GetMetricId()] = append(latest.byMetric[r.GetMetricId()], r)
	}

	return latest
}

// forMetrics returns the latest assessment results of the given metrics. The results are ordered by resource ID and
// metric ID, which is the same order the orchestrator returns them in.
func (l *latestAssessments) forMetrics(metricIds []string) (results []*assessment.AssessmentResult) {
	var (
		seen = make(map[string]struct{}, len(metricIds))
	)

	for _, id := range metricIds {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		results = append(results, l.byMetric[id]...)
	}

	slices.SortStableFunc(results, func(a *assessment.AssessmentResult, b *assessment.AssessmentResult) int {
		if c := strings.Compare(a.GetResourceId(), b.GetResourceId()); c != 0 {
			return c
		}
		return strings.Compare(a.GetMetricId(), b.GetMetricId())
	})

	return results
}
