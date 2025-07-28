package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash                = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleImplementation = errors.New("implementation is not argon2id")
	ErrIncompatibleVersion        = errors.New("incompatible version of argon2")
	ErrMissingParameters          = errors.New("missing a parameter")
	ErrHashMismatch               = errors.New("the password does not match the encoded hash")
)

type Parameters struct {
	Iterations  uint32
	Memory      uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func GenerateRandomBytes(n uint) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func EncodePassword(password string, parameters Parameters) (string, error) {
	salt, saltErr := GenerateRandomBytes(uint(parameters.SaltLength))
	if saltErr != nil {
		return "", saltErr
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		parameters.Iterations,
		parameters.Memory,
		parameters.Parallelism,
		parameters.KeyLength,
	)

	base64Salt := base64.RawStdEncoding.EncodeToString(salt)
	base64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		parameters.Memory,
		parameters.Iterations,
		parameters.Parallelism,
		base64Salt,
		base64Hash,
	)

	return encodedHash, nil
}

func decodeHash(encodedHash string) (parameters Parameters, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		err = ErrInvalidHash
		return
	}

	if vals[1] != "argon2id" {
		err = ErrIncompatibleImplementation
		return
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return
	}

	if version != argon2.Version {
		err = ErrIncompatibleVersion
		return
	}

	parameterPairs := strings.Split(vals[3], ",")
	if len(parameterPairs) != 3 {
		err = ErrMissingParameters
		return
	}

	for _, pair := range parameterPairs {
		var key int
		var value int
		_, err = fmt.Sscanf(pair, "%c=%d", &key, &value)
		if err != nil {
			return
		}

		switch key {
		case 'm':
			parameters.Memory = uint32(value)
		case 't':
			parameters.Iterations = uint32(value)
		case 'p':
			parameters.Parallelism = uint8(value)
		}
	}

	if parameters.Memory == 0 || parameters.Iterations == 0 || parameters.Parallelism == 0 {
		err = ErrMissingParameters
	}

	salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return
	}

	hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return
	}

	parameters.SaltLength = uint32(len(salt))
	parameters.KeyLength = uint32(len(hash))
	return
}

func VerifyPassword(password, encodedHash string) error {
	parameters, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return err
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		parameters.Iterations,
		parameters.Memory,
		parameters.Parallelism,
		parameters.KeyLength,
	)

	if subtle.ConstantTimeCompare(hash, otherHash) != 1 {
		return ErrHashMismatch
	}

	return nil
}
