package main

import (
	"context"
)

type ResourceAuthorizer[R any] struct {
	CanRead   func(context.Context, *R) bool
	CanCreate func(context.Context, *R) bool
	CanUpdate func(context.Context, *R) bool
	CanDelete func(context.Context, *R) bool
}

func AuthorizeResource[I any, O any, R any](
	retrieve func(context.Context, *I) (*R, error),
	handler func(context.Context, *R) (*O, error),
) func(context.Context, *I) (*O, error) {
	return func(ctx context.Context, input *I) (*O, error) {
		resource, err := retrieve(ctx, input)
		if err != nil {
			return nil, err
		}

		return handler(ctx, resource)
	}
}
