package auth

import (
	"errors"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type PayLoad struct {
	ID        string    `json:"id"`
	AccountID int64     `json:"account_id"`
	IssuedAt  time.Time `json:"issued_at"`
	Expiry    time.Time `json:"expiry"`
}

func newPayload(accountID int64, issuedAt, expiry time.Time) *PayLoad {
	return &PayLoad{
		ID:        uuid.NewString(),
		AccountID: accountID,
		IssuedAt:  issuedAt,
		Expiry:    expiry,
	}
}

func CreateToken(accountID int64, issuedAt time.Time, duration time.Duration) (string, *PayLoad, error) {
	expiry := time.Now().UTC().Add(duration)

	payload := newPayload(accountID, issuedAt, expiry)

	token := paseto.NewToken()

	token.SetExpiration(expiry)
	token.SetIssuedAt(payload.IssuedAt)

	token.Set("payload", payload)

	config, err := util.LoadConfig(".")
	if err != nil {
		return "", nil, nil
	}

	secretKey, _ := paseto.NewV4AsymmetricSecretKeyFromHex(config.SecretKeyHex)

	signed := token.V4Sign(secretKey, nil)

	return signed, payload, nil
}

func VerifyToken(signed string) (*PayLoad, error) {
	config, err := util.LoadConfig(".")
	if err != nil {
		return nil, nil
	}

	publicKey, err := paseto.NewV4AsymmetricPublicKeyFromHex(config.PublicKeyHex)
	if err != nil {
		err := errors.New("something went wrong")
		return nil, err
	}

	parser := paseto.NewParser()

	token, err := parser.ParseV4Public(publicKey, signed, nil)
	if err != nil {
		return nil, err
	}

	var payload PayLoad

	if err := token.Get("payload", &payload); err != nil {
		return nil, err
	}

	if time.Now().UTC().After(payload.Expiry) {
		err := errors.New("token is expired")
		return nil, err
	}

	return &payload, nil
}
