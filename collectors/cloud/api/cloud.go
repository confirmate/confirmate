package cloud

import (
	"encoding/json"
	"reflect"

	"confirmate.io/core/api/ontology"
)

// Collector is a part of the discovery service that takes care of the actual discovering and translation into
// ontology objects.
type Collector interface {
	Name() string
	List() ([]ontology.IsResource, error)
	TargetOfEvaluationID() string
}

func Raw(raws ...any) string {
	var rawMap = make(map[string][]any)

	for _, raw := range raws {
		typ := reflect.TypeOf(raw).String()

		rawMap[typ] = append(rawMap[typ], raw)
	}

	b, _ := json.Marshal(rawMap)
	return string(b)
}
