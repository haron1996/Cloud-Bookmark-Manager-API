package paseto

import (
	"encoding/hex"
	"time"

	"github.com/vk-rv/pvx"
)

func TokenGenerator(userID int64, duration time.Time) (string, *Payload, error) {
	payload, err := NewPayload(userID, duration)
	if err != nil {
		return "", payload, err
	}

	k, err := hex.DecodeString("707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f")
	if err != nil {
		return "", payload, err
	}

	symK := pvx.NewSymmetricKey(k, pvx.Version4)

	pv4 := pvx.NewPV4Local()

	token, err := pv4.Encrypt(symK, payload, pvx.WithAssert([]byte("test")))
	if err != nil {
		return "", payload, err
	}

	return token, payload, nil
}

type VerifyTokenChanResStruct struct {
	Payload Payload
	Error   error `json:"error"`
}

func TokenVerifier(token string) (*Payload, error) {
	payload := &Payload{}

	k, err := hex.DecodeString("707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f")
	if err != nil {
		return payload, nil
	}

	symK := pvx.NewSymmetricKey(k, pvx.Version4)

	pv4 := pvx.NewPV4Local()

	err = pv4.Decrypt(token, symK, pvx.WithAssert([]byte("test"))).ScanClaims(payload)
	if err != nil {
		return nil, err
	}

	err = payload.Valid()
	if err != nil {
		switch {
		case err == ErrExpiredToken:
			return nil, err
		default:
			return nil, err
		}
	}

	return payload, nil
}

// implement tenant login
