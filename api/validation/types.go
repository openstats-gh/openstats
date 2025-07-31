package validation

import (
	"encoding/json"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"reflect"
)

type Optional[T any] struct {
	Value    T
	HasValue bool
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &o.Value)
}

// Schema returns a schema representing this value on the wire.
// It returns the schema of the contained type.
func (o *Optional[T]) Schema(r huma.Registry) *huma.Schema {
	return r.Schema(reflect.TypeFor[T](), true, "")
}

func (o *Optional[T]) Receiver() reflect.Value {
	return reflect.ValueOf(o).Elem().Field(0)
}

func (o *Optional[T]) OnParamSet(isSet bool, parsed any) {
	o.HasValue = isSet
}

func (o *Optional[T]) ValueOr(value T) T {
	if o.HasValue {
		return o.Value
	}

	return value
}

type Slug string

func (o *Slug) Schema(r huma.Registry) *huma.Schema {
	return huma.SchemaFromType(r, reflect.TypeFor[string]())
}

type LookupID uuid.UUID

func (l *LookupID) MarshalText() (text []byte, err error) {
	return []byte(uuid.UUID(*l).String()), nil
}

func (l *LookupID) UnmarshalText(text []byte) error {
	parsed, parseErr := uuid.Parse(string(text))
	if parseErr != nil {
		return parseErr
	}

	*l = LookupID(parsed)
	return nil
}

func (l *LookupID) MarshalJSON() (text []byte, err error) {
	return json.Marshal(uuid.UUID(*l).String())
}

func (l *LookupID) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	return l.UnmarshalText([]byte(value))
}

//func (l *LookupID) Schema(r huma.Registry) *huma.Schema {
//	return huma.SchemaFromType(r, reflect.TypeOf(*l))
//	//schema, ok := r.Map()[""]
//	//if !ok {
//	//	schema = &huma.Schema{
//	//
//	//	}
//	//	r.Map()["LookupID"] = schema
//	//}
//	//
//	//return schema
//	//return &huma.Schema{
//	//	Type:                 "LookupID",
//	//	Title:                "LookupID",
//	//	Description:          "",
//	//	Ref:                  "",
//	//	Format:               "",
//	//	ContentEncoding:      "",
//	//	Default:              nil,
//	//	Examples:             nil,
//	//	Items:                nil,
//	//	AdditionalProperties: nil,
//	//	Properties:           nil,
//	//	Enum:                 nil,
//	//	Minimum:              nil,
//	//	ExclusiveMinimum:     nil,
//	//	Maximum:              nil,
//	//	ExclusiveMaximum:     nil,
//	//	MultipleOf:           nil,
//	//	MinLength:            nil,
//	//	MaxLength:            nil,
//	//	Pattern:              "",
//	//	PatternDescription:   "",
//	//	MinItems:             nil,
//	//	MaxItems:             nil,
//	//	UniqueItems:          false,
//	//	Required:             nil,
//	//	MinProperties:        nil,
//	//	MaxProperties:        nil,
//	//	ReadOnly:             false,
//	//	WriteOnly:            false,
//	//	Deprecated:           false,
//	//	Extensions:           nil,
//	//	DependentRequired:    nil,
//	//	OneOf:                nil,
//	//	AnyOf:                nil,
//	//	AllOf:                nil,
//	//	Not:                  nil,
//	//	Discriminator:        nil,
//	//}
//	//return huma.SchemaFromType(r, reflect.TypeFor[uuid.UUID]())
//}
