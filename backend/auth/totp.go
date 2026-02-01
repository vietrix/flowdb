package auth

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTPSecret struct {
	Secret     string
	OtpAuthURL string
}

func GenerateTOTP(accountName string, issuer string) (*TOTPSecret, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Algorithm:   otp.AlgorithmSHA1,
		Period:      30,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return nil, err
	}
	return &TOTPSecret{
		Secret:     key.Secret(),
		OtpAuthURL: key.URL(),
	}, nil
}

func VerifyTOTP(secret string, code string) bool {
	return totp.Validate(code, secret)
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
