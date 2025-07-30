package main

import (
	"context"
	"errors"
	"github.com/dresswithpockets/openstats/app/db/query"
)

var (
	ErrLocalNotFound       = errors.New("local not found")
	ErrLocalNotCorrectType = errors.New("local not correct type")
)

type Local[T any] struct {
	key string
}

func (local *Local[T]) Get(ctx context.Context) (result T, ok bool) {
	localValue := ctx.Value(local.key)
	if localValue == nil {
		ok = false
		return
	}

	result, ok = localValue.(T)
	return
}

func (local *Local[T]) Set(ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, local.key, value)
}

func (local *Local[T]) Exists(ctx context.Context) bool {
	return ctx.Value(local.key) != nil
}

type CtxLocals struct {
	User *Local[*query.User]
}

var Locals = CtxLocals{
	User: &Local[*query.User]{
		key: "User",
	},
}

//func LocalSetUser(c *fiber.Ctx, user *query.User) {
//	c.Locals("user", user)
//}
//
//func LocalGetUser(c *fiber.Ctx) (*models.User, error) {
//	localUser := c.Locals("user")
//	if localUser == nil {
//		return nil, ErrLocalNotFound
//	}
//
//	asUser, asUserOk := localUser.(*models.User)
//	if !asUserOk {
//		return nil, ErrLocalNotCorrectType
//	}
//
//	return asUser, nil
//}
//
//func LocalHasUser(c *fiber.Ctx) error {
//
//}
