package hsm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	domainErr "github.com/LiquidCats/paw/services/litehsm/internal/app/domain/errors"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/hsm"
	"github.com/LiquidCats/paw/services/litehsm/test/assets"
	"github.com/LiquidCats/paw/services/litehsm/test/helpers"
	mocks "github.com/LiquidCats/paw/services/litehsm/test/mocks/litehsm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type KeyManagerSetExpirationSuite struct {
	suite.Suite
}

func TestKeyManagerSetExpirationSuite(t *testing.T) {
	suite.Run(t, new(KeyManagerSetExpirationSuite))
}

func (s *KeyManagerSetExpirationSuite) TestHandleSetsExpiration() {
	keyID := uuid.New()
	expiration := assets.FutureTime()

	repo := mocks.NewMockKeyManagerRepository(s.T())
	repo.EXPECT().
		SetExpiration(
			mock.Anything,
			keyID,
			mock.MatchedBy(func(expiresAt *time.Time) bool {
				return expiresAt != nil && expiresAt.Equal(expiration)
			}),
		).
		Return(nil)

	usecase := hsm.NewKeyManagerSetExpiration(repo)
	err := usecase.Handle(context.Background(), keyID, expiration)

	s.NoError(err)
}

func (s *KeyManagerSetExpirationSuite) TestHandleMapsMissingKeyToDomainError() {
	repo := mocks.NewMockKeyManagerRepository(s.T())
	repo.EXPECT().
		SetExpiration(mock.Anything, mock.Anything, mock.Anything).
		Return(pgx.ErrNoRows)

	usecase := hsm.NewKeyManagerSetExpiration(repo)
	err := usecase.Handle(context.Background(), uuid.New(), assets.FutureTime())

	s.ErrorIs(err, domainErr.ErrKeyNotFound)
}

func (s *KeyManagerSetExpirationSuite) TestHandleWrapsRepositoryError() {
	wantErr := errors.New("update expiration")
	repo := mocks.NewMockKeyManagerRepository(s.T())
	repo.EXPECT().
		SetExpiration(mock.Anything, mock.Anything, mock.Anything).
		Return(wantErr)

	usecase := hsm.NewKeyManagerSetExpiration(repo)
	err := usecase.Handle(context.Background(), uuid.New(), assets.FutureTime())

	s.Require().Error(err)
	s.ErrorIs(err, wantErr)
}

func (s *KeyManagerSetExpirationSuite) TestHandleValidation() {
	past := assets.PastTimeWithHighYearDay()
	tests := []struct {
		name       string
		keyID      entities.KeyID
		expiration time.Time
		wantMsg    string
	}{
		{
			name:       "rejects empty key ID",
			keyID:      entities.KeyID{},
			expiration: assets.FutureTime(),
			wantMsg:    "keyID is required",
		},
		{
			name:       "rejects zero expiration",
			keyID:      uuid.New(),
			expiration: time.Time{},
			wantMsg:    "expiration is required",
		},
		{
			name:       "rejects past expiration",
			keyID:      uuid.New(),
			expiration: past,
			wantMsg:    "expiration date cannot be in the past",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			repo := mocks.NewMockKeyManagerRepository(s.T())
			usecase := hsm.NewKeyManagerSetExpiration(repo)

			err := usecase.Handle(context.Background(), test.keyID, test.expiration)

			helpers.RequireValidationError(s.T(), err, test.wantMsg)
		})
	}
}
