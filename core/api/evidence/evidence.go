// Copyright 2016-2025 Fraunhofer AISEC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package evidence

import (
	"context"
	"reflect"
	"strings"

	"confirmate.io/core/api/ontology"
)

type EvidenceHookFunc func(ctx context.Context, evidence *Evidence, err error)

func (ev *Evidence) GetOntologyResource() ontology.IsResource {
	var (
		resource ontology.IsResource
		ok       bool
	)

	if ev.Resource == nil || ev.Resource.Type == nil {
		return nil
	}

	// A little bit of dark Go magic
	typ := reflect.ValueOf(ev.Resource.Type).Elem()
	resource, ok = typ.Field(0).Interface().(ontology.IsResource)
	if !ok {
		return nil
	}

	return resource
}

// ToResourceSnapshot converts a proto message that complies to the interface [ontology.IsResource]
// into a resource snapshot that can be persisted in our database ([*ResourceSnapshot]).
func ToResourceSnapshot(resource ontology.IsResource, toeId string, toolId string) (r *ResourceSnapshot, err error) {
	// Build a resource snapshot struct. This will hold the latest sync state of the resource for
	// our storage layer.
	r = &ResourceSnapshot{
		Id:                   string(resource.GetId()),
		ResourceType:         strings.Join(ontology.ResourceTypes(resource), ","),
		TargetOfEvaluationId: toeId,
		ToolId:               toolId,
		Resource:             ontology.ProtoResource(resource),
	}

	return
}
