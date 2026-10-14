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


package assessment

import "strings"

// manualResourceIdPrefix is the prefix of the ID of the synthetic resource, to which manually supplied assessment
// results are attached that are about a target of evaluation as a whole and not about a specific resource.
const manualResourceIdPrefix = "manual:"

// ManualTargetOfEvaluationResourceId returns the ID of the synthetic resource, to which manually supplied
// assessment results of the given target of evaluation are attached, if they are not about a specific resource. It is
// unique per target of evaluation, so that the "latest result per resource and metric" of different targets of
// evaluation does not collide.
func ManualTargetOfEvaluationResourceId(targetOfEvaluationId string) string {
	return manualResourceIdPrefix + targetOfEvaluationId
}

// IsManualForTargetOfEvaluation returns true, if the result was supplied manually and is not about a specific
// resource, but about the target of evaluation as a whole.
func (x *AssessmentResult) IsManualForTargetOfEvaluation() bool {
	return x.GetManual() && strings.HasPrefix(x.GetResourceId(), manualResourceIdPrefix)
}
