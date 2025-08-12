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
