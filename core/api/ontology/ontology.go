// Copyright 2016-2025 Fraunhofer AISEC
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

package ontology

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"confirmate.io/core/util"
)

var ErrNotOntologyResource = errors.New("protobuf message is not a valid ontology resource")

type IsResource interface {
	proto.Message
	GetId() string
	GetName() string
	GetCreationTime() *timestamppb.Timestamp
	GetRaw() string
}

type HasRelatedResources interface {
	Related() []string
}

var _ IsResource = &VirtualMachine{}

type Relationship struct {
	Property string
	Value    string
}

func Related(r IsResource) []Relationship {
	var ids []Relationship

	desc := r.ProtoReflect().Descriptor()
	fields := desc.Fields()

	for i := range fields.Len() {
		field := fields.Get(i)

		// We are only interested in string fields
		if field.Kind() != protoreflect.StringKind {
			continue
		}

		// TODO(oxisto): Can we maybe have a proto option on these fields instead of matching by name?
		property, found := strings.CutSuffix(string(field.Name()), "_id")
		if !found {
			// Try with _ids
			property, found = strings.CutSuffix(string(field.Name()), "_ids")
		}
		if found {
			v := r.ProtoReflect().Get(field)
			if field.IsList() {
				list := v.List()
				for i := 0; i < list.Len(); i++ {
					ids = append(ids, Relationship{
						Property: property,
						Value:    list.Get(i).String(),
					})
				}
			} else {
				s := v.String()
				if s != "" {
					ids = append(ids, Relationship{
						Property: property,
						Value:    s,
					})
				}
			}
		}
	}

	return ids
}

func ResourceTypes(r IsResource) []string {
	opts := r.ProtoReflect().Descriptor().Options()

	x := proto.GetExtension(opts, E_ResourceTypeNames)
	if types, ok := x.([]string); ok {
		return types
	}

	return nil
}

// ListResourceTypes returns a list of resource types that are supported by the ontology.
func ListResourceTypes() []string {
	var (
		types    []string
		resource Resource
	)

	// Accessing the descriptor of the resource message.
	md := resource.ProtoReflect().Descriptor()

	// Collect the names of all fields that belong to the oneOf field.
	for i := 0; i < md.Fields().Len(); i++ {
		field := md.Fields().Get(i)
		if field.ContainingOneof() != nil {
			types = append(types, string(field.Name()))
		}
	}

	return types
}

// ResourceMap contains the properties of the resource as a map[string]any, based on its JSON representation.
// This also does some magic to include the resource types in the special key "type".
func ResourceMap(r IsResource) (props map[string]any, err error) {
	var (
		b    []byte
		opts protojson.MarshalOptions
	)

	opts = protojson.MarshalOptions{
		EmitUnpopulated: false, // Only include fields that were explicitly set in the original JSON request, so that policies can distinguish unset fields from default values.
	}

	b, err = opts.Marshal(r)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &props)
	if err != nil {
		return nil, err
	}

	props["type"] = ResourceTypes(r)

	return
}

// ResourceIDs return a list of the given resource IDs
func ResourceIDs(r []IsResource) []string {
	var a = []string{}

	for _, v := range r {
		a = append(a, v.GetId())
	}

	return a
}

// ProtoResource converts a [IsResource] to a [Resource]
func ProtoResource(resource IsResource) *Resource {
	var (
		r  Resource
		m  protoreflect.Message
		od protoreflect.OneofDescriptor
	)

	if util.IsNil(resource) {
		return nil
	}

	// Set up descriptors for proto reflection
	m = r.ProtoReflect()
	od = m.Descriptor().Oneofs().ByName("type")

	// Loop through the fields to find one that matches the resource's protobuf message type
	for i := range od.Fields().Len() {
		field := od.Fields().Get(i)
		if field.Message() == resource.ProtoReflect().Descriptor() {
			m.Set(field, protoreflect.ValueOfMessage(resource.ProtoReflect()))
			break
		}
	}

	return &r
}

// MarshalJSON is a custom JSON marshaller for the [Resource] type that delegates JSON marshalling to the [protojson]
// package.
func (x *Resource) MarshalJSON() (b []byte, err error) {
	return protojson.Marshal(x)
}

// UnmarshalJSON is a custom JSON unmarshaller for the [Resource] type that delegates JSON unmarshalling to the
// [protojson] package.
//
// Resources are persisted as JSON, so stored resources can outlive the version of the ontology they were written
// with. If a stored resource does not match the current ontology, it is read leniently instead of failing the whole
// read, see [migrateLegacyJSON]: unknown fields are dropped and legacy boolean values of properties that became
// lists of strings are converted. The values that are read this way are best-effort only.
func (x *Resource) UnmarshalJSON(b []byte) (err error) {
	var (
		legacy []byte
		strict = protojson.Unmarshal(b, x)
	)

	if strict == nil {
		return nil
	}

	legacy, err = migrateLegacyJSON(b, x.ProtoReflect().Descriptor())
	if err != nil {
		return strict
	}

	proto.Reset(x)
	err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(legacy, x)
	if err != nil {
		return strict
	}

	return nil
}

// migrateLegacyJSON converts the JSON of a message that was written with an older version of the ontology, so that
// it can be read with the current one. A property that is a list of strings in the current ontology, but is still a
// boolean in the given JSON, is converted: true becomes the single value "true" (the information that the property
// was stated is kept, but not what was stated), false becomes an empty list. Other scalars are wrapped into a
// single-element list. Everything that does not match the descriptor is left untouched.
//
// False must not become a value, because the metrics treat a non-empty list as "stated": converting false to
// "false" would turn a property that was explicitly not stated into one that is stated.
func migrateLegacyJSON(b []byte, md protoreflect.MessageDescriptor) (migrated []byte, err error) {
	var (
		v any
	)

	err = json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(migrateLegacyValue(v, md))
}

// migrateLegacyValue walks the decoded JSON object of a message along its descriptor, see [migrateLegacyJSON].
func migrateLegacyValue(v any, md protoreflect.MessageDescriptor) any {
	var (
		obj map[string]any
		fd  protoreflect.FieldDescriptor
		ok  bool
	)

	obj, ok = v.(map[string]any)
	if !ok {
		return v
	}

	for key, value := range obj {
		fd = md.Fields().ByJSONName(key)
		if fd == nil {
			fd = md.Fields().ByName(protoreflect.Name(key))
		}
		if fd == nil || fd.IsMap() {
			continue
		}

		switch {
		case fd.Kind() == protoreflect.StringKind && fd.IsList():
			obj[key] = legacyStringList(value)
		case fd.Kind() == protoreflect.MessageKind && !strings.HasPrefix(string(fd.Message().FullName()), "google.protobuf."):
			if list, isList := value.([]any); isList && fd.IsList() {
				for i := range list {
					list[i] = migrateLegacyValue(list[i], fd.Message())
				}
			} else {
				obj[key] = migrateLegacyValue(value, fd.Message())
			}
		}
	}

	return obj
}

// legacyStringList converts a legacy scalar value of a property that is a list of strings now, see
// [migrateLegacyJSON].
func legacyStringList(v any) any {
	switch value := v.(type) {
	case bool:
		if value {
			return []any{"true"}
		}
		return []any{}
	case string:
		return []any{value}
	case float64:
		return []any{strconv.FormatFloat(value, 'f', -1, 64)}
	default:
		return v
	}
}
