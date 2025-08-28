package auth

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var ValidateOptions = totp.ValidateOpts{
	Period:    30 * 60,
	Digits:    6,
	Algorithm: otp.AlgorithmSHA512,
}
