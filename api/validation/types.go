package validation

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/dresswithpockets/openstats/app/rid"
	"github.com/google/uuid"
	"reflect"
	"strings"
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

type SlugOrRID struct {
	slug string  `hidden:"true"`
	rid  rid.RID `hidden:"true"`
}

func (s *SlugOrRID) Schema(r huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:        "string",
		Title:       "Resource Matcher",
		Description: "A string which can either be a resource's slug or a resource's RID",
		Format:      "resource-matcher",
		Examples: []any{
			"a-valid-slug",
			"u_AZhjuMmhePWkHFALenFEfg",
		},
	}
}

func (s *SlugOrRID) UnmarshalText(text []byte) error {
	textStr := string(text)
	if strings.Contains(textStr, "_") {
		// only RIDs can contain underscores, so this is an RID
		return s.rid.UnmarshalText([]byte(textStr))
	}

	if !ValidSlug(textStr) {
		return errors.New("invalid slug")
	}

	s.slug = textStr
	return nil
}

func (s *SlugOrRID) MarshalText() (text []byte, err error) {
	if s.slug != "" {
		return []byte(s.slug), nil
	}

	return s.rid.MarshalText()
}

func (s *SlugOrRID) Slug() (string, bool) {
	return s.slug, s.slug != ""
}

func (s *SlugOrRID) RID() (rid.RID, bool) {
	return s.rid, s.slug == ""
}

var ErrRidPrefixMismatch = errors.New("rid prefix mismatch")

func EnsureRID(ctx context.Context, s SlugOrRID, prefix string, getUuidBySlug func(context.Context, string) (uuid.UUID, error)) (rid.RID, error) {
	if ridValue, isRid := s.RID(); isRid {
		if ridValue.Prefix != prefix {
			return rid.RID{}, ErrRidPrefixMismatch
		}
		return ridValue, nil
	}

	slugValue, _ := s.Slug()
	uuidValue, getErr := getUuidBySlug(ctx, slugValue)
	if getErr != nil {
		return rid.RID{}, getErr
	}

	return rid.RID{
		Prefix: prefix,
		ID:     uuidValue,
	}, nil
}

//type RID struct {
//	Prefix string
//	UUID   uuid.UUID
//}
//
//var ridSchema = &huma.Schema{
//	Type:            "string<rid>",
//	Nullable:        false,
//	Title:           "RID",
//	Description:     "Resource ID which encodes a Resource Type and a Resource UUID",
//	Ref:             "",      // TODO:
//	Format:          "",      // TODO:
//	ContentEncoding: "utf-8", // TODO:
//	Default:         nil,
//	Examples: []any{
//		"\"u_20W9MCAgSo06z\"",
//	},
//	Items:                nil, // TODO:
//	AdditionalProperties: nil, // TODO:
//	MultipleOf:           nil, // TODO:
//	MinLength:            nil, // TODO:
//	MaxLength:            nil, // TODO:
//	Extensions:           nil,
//}
//
//func (*RID) Schema(huma.Registry) *huma.Schema {
//	return ridSchema
//}

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
