package main

import (
	"errors"
	"github.com/dresswithpockets/openstats/app/query"
	"github.com/gofiber/fiber/v2"
)

var (
	ErrLocalNotFound       = errors.New("local not found")
	ErrLocalNotCorrectType = errors.New("local not correct type")
)

type Local[T any] struct {
	key string
}

func (local *Local[T]) Get(c *fiber.Ctx) (result T, ok bool) {
	localValue := c.Locals(local.key)
	if localValue == nil {
		ok = false
		return
	}

	result, ok = localValue.(T)
	return
}

func (local *Local[T]) Set(c *fiber.Ctx, value T) {
	c.Locals(local.key, value)
}

func (local *Local[T]) Exists(c *fiber.Ctx) bool {
	return c.Locals(local.key) != nil
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
