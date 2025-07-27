package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/fiber/v2/utils"
	sqliteStorage "github.com/gofiber/storage/sqlite3/v2"
	"time"
)

type SessionConfig struct {
	sessionStore *session.Store
}

func (sc *SessionConfig) Get(c *fiber.Ctx) (*Session, error) {
	sess, err := sc.sessionStore.Get(c)
	if err != nil {
		return nil, err
	}

	return &Session{
		session: sess,
	}, nil
}

type Session struct {
	session *session.Session
}

func (s Session) GetUserID() (int32, bool) {
	userId, ok := s.session.Get("UserID").(int32)
	return userId, ok
}

func (s Session) SetUserID(id int32) {
	s.session.Set("UserID", id)
}

func (s Session) Save() error {
	return s.session.Save()
}

func (s Session) Destroy() error {
	return s.session.Destroy()
}

var SessionStore = &SessionConfig{
	sessionStore: session.New(session.Config{
		Expiration:   7 * 24 * time.Hour,
		Storage:      sqliteStorage.New(sqliteStorage.Config{}),
		CookieSecure: true,
		KeyGenerator: utils.UUIDv4,
	}),
}
