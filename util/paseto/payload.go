package paseto

import (
	"errors"
	"sync"
	"time"

	"github.com/kwandapchumba/go-bookmark-manager/util"
)

var (
	ErrExpiredToken = errors.New("token is expired")
	ErrInvalidToken = errors.New("token is invalid")
)

type Payload struct {
	ID         string    `json:"payload_id"`
	UserID     int64     `json:"user_id"`
	ValidUntil time.Time `json:"payload_expiry"`
}

func NewPayload(userID int64, duration time.Time) (*Payload, error) {
	var wg sync.WaitGroup

	stringChan := make(chan string, 1)

	wg.Add(1)

	go func() {
		defer wg.Done()

		util.RandomStringGenerator(stringChan)
	}()

	payload := &Payload{
		ID:         <-stringChan,
		UserID:     userID,
		ValidUntil: duration,
	}

	wg.Wait()

	return payload, nil
}

func (p *Payload) Valid() error {

	if time.Now().UTC().After(p.ValidUntil) {
		return ErrExpiredToken
	}

	return nil
}
