package errors

type ValidationError struct {
	msg string
}

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{msg: msg}
}

func (e *ValidationError) Error() string {
	return e.msg
}

var (
	ErrKeyIsRequired                   = NewValidationError("key is required")
	ErrStatusIsRequired                = NewValidationError("status is required")
	ErrInvalidKeyStatus                = NewValidationError("invalid key status")
	ErrDerivationPathCannotBeSet       = NewValidationError("derivation path cannot be set for key creation")
	ErrKeyIDCannotBeSet                = NewValidationError("key ID cannot be set for key creation")
	ErrSeedFingerprintCannotBeSet      = NewValidationError("seed fingerprint cannot be set for key creation")
	ErrAliasCannotBeEmpty              = NewValidationError("alias cannot be empty")
	ErrAliasCannotBeLessThan3Chars     = NewValidationError("alias cannot be less than 3 characters")
	ErrStatusCannotBeSet               = NewValidationError("status cannot be set for key creation")
	ErrAliasCannotBeLongerThan250Chars = NewValidationError("alias cannot be longer than 250 characters")
	ErrExpirationDateCannotBeInThePast = NewValidationError("expiration date cannot be in the past")
)

type NotFoundError struct {
	msg string
}

func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{msg: msg}
}

func (e *NotFoundError) Error() string {
	return e.msg
}

var ErrKeyNotFound = NewNotFoundError("key not found")
