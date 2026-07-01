package hsm_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	domainErr "github.com/LiquidCats/paw/services/litehsm/internal/app/domain/errors"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/hsm"
	"github.com/LiquidCats/paw/services/litehsm/test/assets"
	mocks "github.com/LiquidCats/paw/services/litehsm/test/mocks/litehsm"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func validKeyEntry(expiresAt *time.Time) entities.KeyEntry {
	return entities.KeyEntry{
		Alias:     "main-key",
		Curve:     entities.CurveTypeSecp256k1,
		Algorithm: entities.AlgorithmTypeECDSA,
		ExpiresAt: expiresAt,
	}
}

type KeyManagerCreateKeySuite struct {
	suite.Suite
}

func TestKeyManagerCreateKeySuite(t *testing.T) {
	suite.Run(t, new(KeyManagerCreateKeySuite))
}

func (s *KeyManagerCreateKeySuite) TestHandleCreatesDisabledKeyWithGeneratedDerivationPath() {
	expiresAt := assets.FutureTime()
	var created entities.KeyEntry

	repo := mocks.NewMockKeyManagerRepository(s.T())
	repo.EXPECT().
		CreateKey(mock.Anything, mock.MatchedBy(func(entry entities.KeyEntry) bool {
			return isCreatedKeyEntry(entry, expiresAt)
		})).
		Run(func(_ context.Context, entry entities.KeyEntry) {
			created = entry
		}).
		Return(nil)

	usecase := hsm.NewKeyManagerCreateKey(repo)
	got, err := usecase.Handle(context.Background(), validKeyEntry(&expiresAt))

	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.True(isCreatedKeyEntry(*got, expiresAt))
	s.Equal(created.DerivationPath.String(), got.DerivationPath.String())
}

func (s *KeyManagerCreateKeySuite) TestHandleWrapsRepositoryError() {
	wantErr := errors.New("insert key")
	repo := mocks.NewMockKeyManagerRepository(s.T())
	repo.EXPECT().
		CreateKey(mock.Anything, mock.Anything).
		Return(wantErr)

	usecase := hsm.NewKeyManagerCreateKey(repo)
	got, err := usecase.Handle(context.Background(), validKeyEntry(nil))

	s.Require().Error(err)
	s.Nil(got)
	s.ErrorIs(err, wantErr)
}

func (s *KeyManagerCreateKeySuite) TestHandleValidation() {
	past := assets.PastTimeWithHighYearDay()
	tests := []struct {
		name    string
		mutate  func(*entities.KeyEntry)
		wantErr error
	}{
		{
			name: "rejects provided derivation path",
			mutate: func(entry *entities.KeyEntry) {
				entry.DerivationPath = entities.DerivationPath{entities.NewIndex(0, false)}
			},
			wantErr: domainErr.ErrDerivationPathCannotBeSet,
		},
		{
			name: "rejects provided key ID",
			mutate: func(entry *entities.KeyEntry) {
				entry.KeyID = uuid.New()
			},
			wantErr: domainErr.ErrKeyIDCannotBeSet,
		},
		{
			name: "rejects provided seed fingerprint",
			mutate: func(entry *entities.KeyEntry) {
				entry.SeedFingerprint = "seed"
			},
			wantErr: domainErr.ErrSeedFingerprintCannotBeSet,
		},
		{
			name: "rejects empty alias",
			mutate: func(entry *entities.KeyEntry) {
				entry.Alias = ""
			},
			wantErr: domainErr.ErrAliasCannotBeEmpty,
		},
		{
			name: "rejects short alias",
			mutate: func(entry *entities.KeyEntry) {
				entry.Alias = "ab"
			},
			wantErr: domainErr.ErrAliasCannotBeLessThan3Chars,
		},
		{
			name: "rejects provided status",
			mutate: func(entry *entities.KeyEntry) {
				entry.Status = entities.KeyStatusEnabled
			},
			wantErr: domainErr.ErrStatusCannotBeSet,
		},
		{
			name: "rejects long alias",
			mutate: func(entry *entities.KeyEntry) {
				entry.Alias = strings.Repeat("a", 251)
			},
			wantErr: domainErr.ErrAliasCannotBeLongerThan250Chars,
		},
		{
			name: "rejects past expiration",
			mutate: func(entry *entities.KeyEntry) {
				entry.ExpiresAt = &past
			},
			wantErr: domainErr.ErrExpirationDateCannotBeInThePast,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			repo := mocks.NewMockKeyManagerRepository(s.T())
			usecase := hsm.NewKeyManagerCreateKey(repo)
			entry := validKeyEntry(nil)
			test.mutate(&entry)

			got, err := usecase.Handle(context.Background(), entry)

			s.Require().Error(err)
			s.Nil(got)
			s.ErrorIs(err, test.wantErr)
		})
	}
}

func isCreatedKeyEntry(entry entities.KeyEntry, expiresAt time.Time) bool {
	return entry.Alias == "main-key" &&
		entry.Curve == entities.CurveTypeSecp256k1 &&
		entry.Algorithm == entities.AlgorithmTypeECDSA &&
		entry.Status == entities.KeyStatusDisabled &&
		entry.ExpiresAt != nil &&
		entry.ExpiresAt.Equal(expiresAt) &&
		len(entry.DerivationPath) == 3 &&
		entry.DerivationPath[0].IsHardened() &&
		entry.DerivationPath[1].IsHardened() &&
		!entry.DerivationPath[2].IsHardened() &&
		entry.DerivationPath[2].Uint32() == 0
}
