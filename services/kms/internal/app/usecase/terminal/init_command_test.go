package terminal_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/ports"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/usecase/terminal"
	mocks "github.com/LiquidCats/paw/services/litehsm/test/mocks/litehsm"
	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// newSealedEnvelope builds a fully populated *entities.Envelope so MarshalBinary
// and Destroy operate on real locked buffers, mirroring what the sealer returns.
// A cleanup is registered so the buffers are freed if the test aborts before
// RunArgs has a chance to call env.Destroy().
func newSealedEnvelope(t *testing.T) *entities.Envelope {
	t.Helper()
	env := &entities.Envelope{
		Magic:      [4]byte{'H', 'S', 'M', '1'},
		Version:    entities.DefaultVersion,
		KDF:        entities.KDFArgon2id,
		AEAD:       entities.AEADxchacha20poly1305,
		KDFParams:  entities.DefaultKDFParams,
		Salt:       memguard.NewBuffer(entities.SaltSize),
		NonceKEK:   memguard.NewBuffer(entities.NonceSize),
		NonceDEK:   memguard.NewBuffer(entities.NonceSize),
		WrappedDEK: memguard.NewBuffer(entities.WrappedDEKSize),
		Ciphertext: memguard.NewBufferFromBytes([]byte("sealed-seed-payload")),
	}
	t.Cleanup(func() {
		if env.Salt != nil && env.Salt.IsAlive() {
			env.Destroy()
		}
	})
	return env
}

// ── suite ────────────────────────────────────────────────────────────────────

type InitCommandSuite struct {
	suite.Suite
}

func TestInitCommandSuite(t *testing.T) {
	suite.Run(t, new(InitCommandSuite))
}

func (s *InitCommandSuite) mustNewCommand(sealer ports.Sealer, providers ...ports.PassphraseProvider) *terminal.InitCommand {
	s.T().Helper()
	uc, err := terminal.NewInitCommand(sealer, providers...)
	s.Require().NoError(err)
	return uc
}

// openBytes opens an enclave and returns a clone of its plaintext bytes.
func (s *InitCommandSuite) openBytes(enc *memguard.Enclave) []byte {
	s.T().Helper()
	lb, err := enc.Open()
	s.Require().NoError(err)
	defer lb.Destroy()
	return bytes.Clone(lb.Bytes())
}

// ── constructor ──────────────────────────────────────────────────────────────

func (s *InitCommandSuite) TestNewInitCommandValidatesDependencies() {
	s.Run("nil sealer", func() {
		_, err := terminal.NewInitCommand(nil, mocks.NewMockPassphraseProvider(s.T()))
		s.Error(err)
	})
	s.Run("no providers", func() {
		_, err := terminal.NewInitCommand(mocks.NewMockSealer(s.T()))
		s.Error(err)
	})
	s.Run("nil provider", func() {
		_, err := terminal.NewInitCommand(mocks.NewMockSealer(s.T()), nil)
		s.Error(err)
	})
}

// ── RunArgs ──────────────────────────────────────────────────────────────────

func (s *InitCommandSuite) TestRunUsesParsedFlagsAndWritesEnvelope() {
	output := filepath.Join(s.T().TempDir(), "seal.bin")
	envelope := newSealedEnvelope(s.T())

	expected, err := envelope.MarshalBinary()
	s.Require().NoError(err)

	// Name() is called several times inside RunArgs (loop, default flag value,
	// extractPassphrase walk). Registering without Times() matches all calls.
	stdinProvider := mocks.NewMockPassphraseProvider(s.T())
	stdinProvider.EXPECT().Name().Return("stdin")

	envProvider := mocks.NewMockPassphraseProvider(s.T())
	envProvider.EXPECT().Name().Return("env")
	envProvider.EXPECT().Get("HSM_PASSPHRASE").
		Return(memguard.NewBufferFromBytes([]byte("secret")), nil)

	var seedBuf *memguard.LockedBuffer
	var capturedPassphrase []byte
	var capturedSeedSize int

	sealer := mocks.NewMockSealer(s.T())
	sealer.EXPECT().Seal(mock.Anything, mock.Anything).
		Run(func(passphrase *memguard.Enclave, data *memguard.LockedBuffer) {
			capturedPassphrase = s.openBytes(passphrase)
			capturedSeedSize = data.Size()
			seedBuf = data
		}).
		Return(envelope, nil)

	uc := s.mustNewCommand(sealer, stdinProvider, envProvider)
	s.Require().NoError(uc.RunArgs([]string{
		"-from", "env",
		"-input", "HSM_PASSPHRASE",
		"-output", output,
	}))

	got, err := os.ReadFile(output)
	s.Require().NoError(err)
	s.Equal(expected, got, "output file content must match sealed envelope")

	info, err := os.Stat(output)
	s.Require().NoError(err)
	s.Equal(os.FileMode(0o600), info.Mode().Perm(), "output file must be owner-read-write only")

	s.Equal([]byte("secret"), capturedPassphrase, "sealer must receive the passphrase bytes")
	s.Equal(32, capturedSeedSize, "seed must be 32 random bytes")
	s.False(seedBuf.IsAlive(), "seed buffer must be destroyed after run")
	s.False(envelope.Salt.IsAlive(), "envelope must be destroyed after run")
}

func (s *InitCommandSuite) TestRunRequiresOutputFlag() {
	provider := mocks.NewMockPassphraseProvider(s.T())
	provider.EXPECT().Name().Return("env")
	// No Get expectation: if Get is called the mock panics, enforcing the
	// invariant that the passphrase is never read before the output is validated.

	uc := s.mustNewCommand(mocks.NewMockSealer(s.T()), provider)
	err := uc.RunArgs([]string{"-from", "env", "-input", "HSM_PASSPHRASE"})

	s.Require().ErrorContains(err, "output path is required")
}

func (s *InitCommandSuite) TestRunReturnsFlagParseError() {
	provider := mocks.NewMockPassphraseProvider(s.T())
	provider.EXPECT().Name().Return("env")

	uc := s.mustNewCommand(mocks.NewMockSealer(s.T()), provider)
	err := uc.RunArgs([]string{"-unknown-flag"})

	s.Require().ErrorContains(err, "init args")
}

func (s *InitCommandSuite) TestRunRejectsNilPassphrase() {
	output := filepath.Join(s.T().TempDir(), "seal.bin")
	provider := mocks.NewMockPassphraseProvider(s.T())
	provider.EXPECT().Name().Return("env")
	provider.EXPECT().Get(mock.Anything).Return(nil, nil)
	// No Seal expectation: if Seal is called the mock panics.

	uc := s.mustNewCommand(mocks.NewMockSealer(s.T()), provider)
	err := uc.RunArgs([]string{"-from", "env", "-output", output})

	s.Require().ErrorContains(err, "passphrase not found")
}

func (s *InitCommandSuite) TestRunRejectsNilEnvelope() {
	output := filepath.Join(s.T().TempDir(), "seal.bin")
	provider := mocks.NewMockPassphraseProvider(s.T())
	provider.EXPECT().Name().Return("env")
	provider.EXPECT().Get(mock.Anything).
		Return(memguard.NewBufferFromBytes([]byte("secret")), nil)

	var seedBuf *memguard.LockedBuffer
	sealer := mocks.NewMockSealer(s.T())
	sealer.EXPECT().Seal(mock.Anything, mock.Anything).
		Run(func(_ *memguard.Enclave, data *memguard.LockedBuffer) {
			seedBuf = data
		}).
		Return(nil, nil)

	uc := s.mustNewCommand(sealer, provider)
	err := uc.RunArgs([]string{"-from", "env", "-output", output})

	s.Require().ErrorContains(err, "sealer returned nil envelope")
	s.False(seedBuf.IsAlive(), "seed buffer must be destroyed even on nil envelope")
}

func (s *InitCommandSuite) TestRunRejectsUnsupportedPassphraseSource() {
	output := filepath.Join(s.T().TempDir(), "seal.bin")
	provider := mocks.NewMockPassphraseProvider(s.T())
	provider.EXPECT().Name().Return("env")
	// No Get expectation: unsupported source is rejected before Get is called.

	uc := s.mustNewCommand(mocks.NewMockSealer(s.T()), provider)
	err := uc.RunArgs([]string{"-from", "file", "-output", output})

	s.Require().ErrorContains(err, `unsupported source type "file"`)
}
