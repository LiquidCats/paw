package hsm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	domainErr "github.com/LiquidCats/paw/services/litehsm/internal/app/domain/errors"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/hsm"
	mocks "github.com/LiquidCats/paw/services/litehsm/test/mocks/litehsm"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type KeyManagerSetStatusSuite struct {
	suite.Suite
}

func TestKeyManagerSetStatusSuite(t *testing.T) {
	suite.Run(t, new(KeyManagerSetStatusSuite))
}

func (s *KeyManagerSetStatusSuite) TestHandleSetsStatus() {
	keyID := uuid.New()
	repo := mocks.NewMockKeyManagerRepository(s.T())
	repo.EXPECT().
		SetStatus(mock.Anything, keyID, entities.KeyStatusEnabled).
		Return(nil)

	usecase := hsm.NewKeyManagerSetStatus(repo)
	err := usecase.Handle(context.Background(), keyID, entities.KeyStatusEnabled)

	s.NoError(err)
}

func (s *KeyManagerSetStatusSuite) TestHandleMapsMissingKeyToDomainError() {
	repo := mocks.NewMockKeyManagerRepository(s.T())
	repo.EXPECT().
		SetStatus(mock.Anything, mock.Anything, entities.KeyStatusDisabled).
		Return(pgx.ErrNoRows)

	usecase := hsm.NewKeyManagerSetStatus(repo)
	err := usecase.Handle(context.Background(), uuid.New(), entities.KeyStatusDisabled)

	s.ErrorIs(err, domainErr.ErrKeyNotFound)
}

func (s *KeyManagerSetStatusSuite) TestHandleWrapsRepositoryError() {
	wantErr := errors.New("update status")
	repo := mocks.NewMockKeyManagerRepository(s.T())
	repo.EXPECT().
		SetStatus(mock.Anything, mock.Anything, entities.KeyStatusDeleted).
		Return(wantErr)

	usecase := hsm.NewKeyManagerSetStatus(repo)
	err := usecase.Handle(context.Background(), uuid.New(), entities.KeyStatusDeleted)

	s.Require().Error(err)
	s.ErrorIs(err, wantErr)
}

func (s *KeyManagerSetStatusSuite) TestHandleValidation() {
	tests := []struct {
		name    string
		keyID   entities.KeyID
		status  entities.KeyStatus
		wantErr error
	}{
		{
			name:    "rejects empty key ID",
			keyID:   entities.KeyID{},
			status:  entities.KeyStatusEnabled,
			wantErr: domainErr.ErrKeyIsRequired,
		},
		{
			name:    "rejects empty status",
			keyID:   uuid.New(),
			status:  "",
			wantErr: domainErr.ErrStatusIsRequired,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			repo := mocks.NewMockKeyManagerRepository(s.T())
			usecase := hsm.NewKeyManagerSetStatus(repo)

			err := usecase.Handle(context.Background(), test.keyID, test.status)

			s.Require().Error(err)
			s.ErrorIs(err, test.wantErr)
		})
	}
}
