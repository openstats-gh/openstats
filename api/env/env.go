package env

import (
	"errors"
	"github.com/joho/godotenv"
	"github.com/rotisserie/eris"
	"io/fs"
	"log"
	"maps"
	"os"
	"slices"
	"strings"
)

func GetBool(key string) bool {
	value := os.Getenv(key)
	return value == "true" || value == "1"
}

func GetString(key string) string {
	return os.Getenv(key)
}

func GetMapped[T any](key string, into map[string]T) (T, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		var result T
		return result, eris.Errorf("%s must be set", key)
	}

	result, ok := into[value]
	if !ok {
		validValues := strings.Join(slices.Collect(maps.Keys(into)), ", ")
		return result, eris.Errorf("%s has value '%s', expected one these values: %s", key, value, validValues)
	}

	return result, nil
}

func GetList(key string) []string {
	return strings.Split(os.Getenv(key), ",")
}

func Require(keys ...string) {
	var missing []string
	for _, key := range keys {
		if _, exists := os.LookupEnv(key); !exists {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		log.Fatalf("Missing required environment variables: %s", strings.Join(missing, ", "))
	}
}

// Load loads .env files, using the OPENSTATS_ENV envvar to determine which .env files to load.
//
// If OPENSTATS_ENV is empty or not set, it defaults to development.
//
// The .env precedence is as follows:
//
//	.env.{env}.local
//	.env.local
//	.env.{env}
//	.env
//
// If OPENSTATS_ENV is "test", .env.local is not loaded.
func Load() (err error) {
	env := os.Getenv("OPENSTATS_ENV")
	if env == "" {
		env = "development"
	}

	err = godotenv.Load(".env." + env + ".local")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return
	}

	if env != "test" {
		err = godotenv.Load(".env.local")
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return
		}
	}

	err = godotenv.Load(".env." + env)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return
	}

	err = godotenv.Load(".env")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return
	}

	return nil
}
