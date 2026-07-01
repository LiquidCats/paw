package entities

import (
	"slices"
	"time"

	domainErr "github.com/LiquidCats/paw/services/litehsm/internal/app/domain/errors"
	v1 "github.com/LiquidCats/paw/protos/gen/go/services/litehsm/v1"
	"github.com/google/uuid"
)

type KeyStatus string

const (
	KeyStatusEnabled  KeyStatus = "enabled"
	KeyStatusDisabled KeyStatus = "disabled"
	KeyStatusDeleted  KeyStatus = "deleted"
)

func (k KeyStatus) Can(n KeyStatus) bool {
	switch k {
	case KeyStatusEnabled:
		return n.OneOf(KeyStatusDeleted, KeyStatusDisabled)
	case KeyStatusDisabled:
		return n.OneOf(KeyStatusDeleted, KeyStatusEnabled)
	case KeyStatusDeleted:
		return false
	default:
		return false
	}
}

func (k KeyStatus) OneOf(n ...KeyStatus) bool {
	return slices.Contains(n, k)
}

func KeyStatusFromProto(s v1.KeyStatus) (KeyStatus, error) {
	switch s { //nolint:exhaustive
	case v1.KeyStatus_KEY_STATUS_ENABLED:
		return KeyStatusEnabled, nil
	case v1.KeyStatus_KEY_STATUS_DISABLED:
		return KeyStatusDisabled, nil
	case v1.KeyStatus_KEY_STATUS_DELETED:
		return KeyStatusDeleted, nil
	default:
		return "", domainErr.ErrInvalidKeyStatus
	}
}

func (k KeyStatus) ToProto() v1.KeyStatus {
	switch k {
	case KeyStatusEnabled:
		return v1.KeyStatus_KEY_STATUS_ENABLED
	case KeyStatusDisabled:
		return v1.KeyStatus_KEY_STATUS_DISABLED
	case KeyStatusDeleted:
		return v1.KeyStatus_KEY_STATUS_DELETED
	default:
		return v1.KeyStatus_KEY_STATUS_UNSPECIFIED
	}
}

type KeyEntry struct {
	KeyID           KeyID
	SeedFingerprint string
	Alias           string
	Curve           CurveType
	Algorithm       AlgorithmType
	DerivationPath  DerivationPath
	Status          KeyStatus
	ExpiresAt       *time.Time
}

type KeyID = uuid.UUID
