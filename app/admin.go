package main

import (
	"errors"
	"github.com/gofiber/fiber/v2/log"
)

const (
	RootUserDisplayName = "Admin"
	RootUserEmail       = ""
	RootUserSlug        = "openstats"
	RootUserPass        = "openstatsadmin"
)

func AddRootAdminUser() {
	_, newUserErr := AddNewUser(RootUserDisplayName, RootUserEmail, RootUserSlug, RootUserPass)
	// this function is expected to be idempotent - if called multiple times, it shouldn't fail even if the admin
	// already exists
	if newUserErr != nil && !errors.Is(newUserErr, ErrSlugAlreadyInUse) {
		log.Fatal(newUserErr)
	}
}

func IsAdmin(user *User) bool {
	// TODO: add distinction between Admin and Root - Root should be able to add non-root Admin users, which can do
	//       everything except add other Admins
	return user != nil && user.Slug == RootUserSlug
}

func IsRoot(user *User) bool {
	return user != nil && user.Slug == RootUserSlug
}
