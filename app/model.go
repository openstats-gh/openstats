package main

import (
	"database/sql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"time"
)

type MutableKeyedModel struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ImmutableKeyedModel struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type MutableKeylessModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ImmutableKeylessModel struct {
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type User struct {
	MutableKeyedModel
	Slug            string `gorm:"unique"`
	PasswordId      uint64 `gorm:"<-:create;unique"`
	Password        UserPassword
	SlugRecords     []UserSlugRecord
	Emails          []UserEmail
	DisplayNames    []UserDisplayName
	Developers      []Developer `gorm:"many2many:developer_members"`
	ProgressRecords []AchievementProgress
}

type UserPassword struct {
	MutableKeyedModel
	UserID      uint64 `gorm:"<-:create"`
	EncodedHash string
}

type UserSlugRecord struct {
	ImmutableKeyedModel
	UserID uint64 `gorm:"<-:create"`
	Slug   string `gorm:"<-:create"`
}

type UserEmail struct {
	MutableKeyedModel
	UserID      uint64       `gorm:"<-:create"`
	Email       string       `gorm:"<-:create;index:user_email"`
	ConfirmedAt sql.NullTime `gorm:"index"`
}

type UserDisplayName struct {
	ImmutableKeyedModel
	UserID uint64 `gorm:"<-:create"`
	Name   string `gorm:"<-:create"`
}

type Developer struct {
	MutableKeyedModel
	Slug        string `gorm:"unique"`
	SlugRecords []DeveloperSlugRecord
	Members     []User `gorm:"many2many:developer_members"`
	Games       []Game
}

type DeveloperMember struct {
	ImmutableKeylessModel
	UserID      uint64 `gorm:"<-:create;primaryKey"`
	DeveloperID uint64 `gorm:"<-:create;primaryKey"`
}

func (DeveloperMember) BeforeCreate(db *gorm.DB) error {
	return nil
}

type DeveloperSlugRecord struct {
	ImmutableKeyedModel
	DeveloperID uint64 `gorm:"<-:create"`
	Slug        string `gorm:"<-:create"`
}

type Game struct {
	MutableKeyedModel
	DeveloperID  uint64 `gorm:"<-:create;index:idx_game_dev_slug,unique"`
	Slug         string `gorm:"index:idx_game_dev_slug,unique"`
	Achievements []Achievement
}

type Achievement struct {
	MutableKeyedModel
	GameID              uint64 `gorm:"<-:create;index:idx_game_achievement_slug,unique"`
	Slug                string `gorm:"index:idx_game_achievement_slug,unique"`
	Name                string
	Description         string
	ProgressRequirement *uint64
	ProgressRecords     []AchievementProgress
}

type AchievementProgress struct {
	MutableKeyedModel
	UserID        uint64 `gorm:"<-:create;primaryKey;autoIncrement:false"`
	AchievementID uint64 `gorm:"<-:create;primaryKey;autoIncrement:false"`
	Progress      uint64 `gorm:"index"`
}

var DB *gorm.DB

func SetupDB() error {
	db, dbErr := gorm.Open(sqlite.Open("openstats.db"), &gorm.Config{
		TranslateError: true,
	})
	if dbErr != nil {
		return dbErr
	}

	setupErr := db.SetupJoinTable(&User{}, "Developers", &DeveloperMember{})
	if setupErr != nil {
		return setupErr
	}

	setupErr = db.SetupJoinTable(&Developer{}, "Members", &DeveloperMember{})
	if setupErr != nil {
		return setupErr
	}

	migrateErr := db.AutoMigrate(
		&User{},
		&UserPassword{},
		&UserEmail{},
		&UserSlugRecord{},
		&UserDisplayName{},
		&Developer{},
		&DeveloperSlugRecord{},
		&Game{},
		&Achievement{},
		&AchievementProgress{},
	)
	if migrateErr != nil {
		return migrateErr
	}

	DB = db
	return nil
}
