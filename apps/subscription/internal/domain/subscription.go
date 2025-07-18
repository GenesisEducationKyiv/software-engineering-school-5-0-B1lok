package domain

import (
	"net/mail"
	"time"

	internalErrors "subscription-service/internal/errors"
	pkgErrors "subscription-service/pkg/errors"

	"github.com/google/uuid"
)

type Frequency string

const (
	FrequencyHourly Frequency = "hourly"
	FrequencyDaily  Frequency = "daily"
)

type Subscription struct {
	ID        uint
	Email     string
	City      string
	Frequency Frequency
	Token     string
	Confirmed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SubscriptionLookup struct {
	Email     string
	City      string
	Frequency Frequency
}

type GroupedSubscription struct {
	City          string
	Subscriptions []*Subscription
}

func NewSubscription(email, city string, frequency Frequency) (*Subscription, error) {
	sub := &Subscription{
		Email:     email,
		City:      city,
		Frequency: frequency,
		Token:     uuid.New().String(),
		Confirmed: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := sub.validate(); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *Subscription) SetConfirmed(confirmed bool) {
	s.Confirmed = confirmed
	s.UpdatedAt = time.Now()
}

func (s *Subscription) validate() error {
	if _, err := mail.ParseAddress(s.Email); err != nil {
		return pkgErrors.New(internalErrors.ErrInvalidInput, "invalid email address")
	}

	if s.City == "" {
		return pkgErrors.New(internalErrors.ErrInvalidInput, "city is required")
	}

	if !isValidFrequency(s.Frequency) {
		return pkgErrors.New(internalErrors.ErrInvalidInput, "invalid frequency value")
	}

	return nil
}

func isValidFrequency(f Frequency) bool {
	return f == FrequencyHourly || f == FrequencyDaily
}
