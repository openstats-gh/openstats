package auth

import (
	"context"
	"errors"
	"github.com/dresswithpockets/openstats/app/db"
	"github.com/dresswithpockets/openstats/app/db/query"
	"github.com/gofiber/fiber/v2/log"
)

const (
	RootUserDisplayName = "Admin"
	RootUserEmail       = ""
	RootUserSlug        = "openstats"
	RootUserPass        = "openstatsadmin"
)

func AddRootAdminUser(ctx context.Context) {
	_, newUserErr := AddNewUser(ctx, RootUserDisplayName, RootUserEmail, RootUserSlug, RootUserPass)
	// this function is expected to be idempotent - if called multiple times, it shouldn't fail even if the admin
	// already exists
	if newUserErr != nil && !errors.Is(newUserErr, db.ErrSlugAlreadyInUse) {
		log.Fatal(newUserErr)
	}
}

func IsAdmin(user query.User) bool {
	// TODO: add distinction between Admin and Root - Root should be able to add non-root Admin users, which can do
	//       everything except add other Admins
	return user.Slug == RootUserSlug
}

// IsRoot returns true if the user is determined to have Root privileges
//
//goland:noinspection GoUnusedExportedFunction
func IsRoot(user query.User) bool {
	return user.Slug == RootUserSlug
}
