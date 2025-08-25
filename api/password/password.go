package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"github.com/rotisserie/eris"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash                = eris.New("the encoded hash is not in the correct format")
	ErrIncompatibleImplementation = eris.New("implementation is not argon2id")
	ErrCantDecodeVersion          = eris.New("can't decode the argon2id version")
	ErrIncompatibleVersion        = eris.New("incompatible version of argon2")
	ErrMissingParameters          = eris.New("missing a parameter")
	ErrHashMismatch               = eris.New("the password does not match the encoded hash")
)

type Parameters struct {
	Iterations  uint32
	Memory      uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// GenerateRandomBytes generates n random bytes into a []byte from a cryptographic stream
//
// err will always be wrapped with eris.Wrap
func GenerateRandomBytes(n uint) (b []byte, err error) {
	b = make([]byte, n)
	_, err = rand.Read(b)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate random bytes")
	}

	return
}

// EncodePassword encodes the given password & parameters into a string; the algorithm, parameters, salt, and hash can
// be decoded from this string and later used to validate any given password string.
//
// Any errors will always be wrapped by eris.Wrap
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
		err = eris.Wrap(err, "error decoding implementation version")
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
			err = eris.Wrap(err, "error decoding a parameter")
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
		err = eris.Wrap(err, "error decoding salt bytes from base64")
		return
	}

	hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		err = eris.Wrap(err, "error decoding hash bytes from base64")
		return
	}

	parameters.SaltLength = uint32(len(salt))
	parameters.KeyLength = uint32(len(hash))
	return
}

// VerifyPassword returns a nil error if the password passes validation against the encoded password information
// in encodedHash
//
// Any errors will always be wrapped by eris.Wrap
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
