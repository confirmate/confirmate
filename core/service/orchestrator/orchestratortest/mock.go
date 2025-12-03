package orchestratortest

import "confirmate.io/core/api/assessment"

var (
	MockMetric1 = &assessment.Metric{
		Id:          "metric-1",
		Description: "Mock Metric 1",
	}
	MockMetric2 = &assessment.Metric{
		Id:          "metric-2",
		Description: "Mock Metric 2",
	}
)
