package collector

import (
	"encoding/json"
	"reflect"

	"confirmate.io/core/api/ontology"
)

// Collector takes care of collecting provider resources and translating them into ontology objects.
type Collector interface {
	Name() string
	ID() string
	Collect() ([]ontology.IsResource, error)
	List() ([]ontology.IsResource, error)
	TargetOfEvaluationID() string
}

// Raw serializes provider-native objects into the ontology raw payload field.
func Raw(raws ...any) string {
	var rawMap = make(map[string][]any)

	for _, raw := range raws {
		typ := reflect.TypeOf(raw).String()

		rawMap[typ] = append(rawMap[typ], raw)
	}

	b, _ := json.Marshal(rawMap)
	return string(b)
}
