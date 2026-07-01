# litehsm

`litehsm` is a Go key-management service for deterministic secp256k1 key material. It exposes a gRPC API for creating and managing key metadata, stores key records in PostgreSQL, and derives key material from a sealed root seed held in locked memory.

The repository also includes `litehsmctl`, a small CLI used to initialize the sealed seed envelope consumed by the service at startup.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Security Model](#security-model)
- [Configuration](#configuration)
- [Seed Initialization](#seed-initialization)
- [gRPC API](#grpc-api)
- [Database](#database)
- [Running Locally](#running-locally)
- [Docker](#docker)
- [Development](#development)

---

## Overview

At startup, `litehsm`:

1. Loads configuration from environment variables with the `LITEHSM` prefix.
2. Clears the process environment after configuration is read.
3. Connects to PostgreSQL through `pgxpool`.
4. Runs embedded database migrations.
5. Reads a sealed seed envelope from configured sensitive input.
6. Reads the seed passphrase from configured sensitive input.
7. Unseals the seed with XChaCha20-Poly1305 and Argon2id-backed key derivation.
8. Builds a BIP-32 secp256k1 master keychain from the unsealed seed.
9. Starts the gRPC key-manager service.

Key creation currently persists metadata and a generated derivation path. The startup keychain wiring is present, while key-derivation-backed signing/export operations are not exposed by the current gRPC server.

---

## Architecture

The codebase follows a ports-and-adapters layout.

```text
cmd/
├── litehsm/                 # service entrypoint
└── litehsmctl/              # seed initialization CLI

configs/                     # envconfig-backed configuration structs

internal/
├── adapter/
│   ├── keychain/            # BIP-32 secp256k1 keychain implementation
│   ├── passphrase/          # stdin/env/file passphrase providers
│   ├── postgresql/          # repository, transactions, sqlc, migrations
│   ├── sealer/              # seed envelope sealing/unsealing
│   └── transport/grpc/      # gRPC KeyManagerService adapter
├── app/
│   ├── domain/              # entities and domain errors
│   ├── ports/               # repository, sealer, keychain, provider interfaces
│   └── usecase/
│       ├── hsm/             # create key, set expiration, set status
│       └── terminal/        # litehsmctl init command
└── bootstrap/               # app name and envelope magic constants

pkg/                         # base58, hash160, keccak, sha, config helpers

test/                        # test assets, helpers, generated mocks
```

### Runtime Data Flow

```text
Environment
  → configs.Load("litehsm")
  → PostgreSQL pool + embedded migrations
  → seed envelope + passphrase sensitive params
  → sealer.Unseal(...)
  → keychain.NewSecp256k1Keychain(...)
  → gRPC KeyManagerService
```

### Key Creation Flow

```text
gRPC CreateKey request
  → protobuf curve/algorithm conversion
  → KeyManagerCreateKey use case
  → validation
  → random hardened derivation path: m/<random>'/<random>'/0
  → disabled key metadata
  → PostgreSQL repository
```

---

## Security Model

Sensitive material is handled with [`memguard`](https://github.com/awnumar/memguard):

- Seed, passphrases, and decrypted buffers live in locked memory where supported.
- Buffers are wiped when destroyed.
- `memguard.CatchInterrupt()` and `memguard.Purge()` are used by the service process.
- Configuration loading calls `os.Clearenv()` after env vars are processed.
- The seed is stored on disk as a sealed envelope, not plaintext.

The seed envelope uses the project envelope format with magic `HSM1`, Argon2id KDF parameters, and XChaCha20-Poly1305 encryption.

---

## Configuration

Configuration is loaded from environment variables using [`envconfig`](https://github.com/kelseyhightower/envconfig) with the `LITEHSM` prefix. YAML tags exist on config structs, but the current loader does not read a YAML config file.

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `LITEHSM_APP_KEYMANAGER_SEED_PASSPHRASE` | Sensitive source for the seed-unsealing passphrase | required |
| `LITEHSM_APP_KEYMANAGER_SEED_SEALING` | Sensitive source for the sealed seed envelope | required |
| `LITEHSM_COMMON_GRPC_PORT` | gRPC listen port | `50051` |
| `LITEHSM_COMMON_GRPC_CONN_TIMEOUT` | gRPC connection timeout | `120s` |
| `LITEHSM_COMMON_LOGGING_LEVEL` | `slog` log level | `info` |
| `LITEHSM_COMMON_DB_DRIVER` | Database URL scheme | `postgres` |
| `LITEHSM_COMMON_DB_HOST` | PostgreSQL host | required |
| `LITEHSM_COMMON_DB_PORT` | PostgreSQL port | required |
| `LITEHSM_COMMON_DB_DATABASE` | PostgreSQL database name | required |
| `LITEHSM_COMMON_DB_USER` | PostgreSQL user | required |
| `LITEHSM_COMMON_DB_PASSWORD` | PostgreSQL password | required |
| `LITEHSM_COMMON_DB_SSL` | Enables `sslmode=enable` when true | `false` |

The PostgreSQL DSN is built as:

```text
<driver>://<user>:<password>@<host>:<port>/<database>?sslmode=<disable|enable>
```

### Sensitive Parameters

Sensitive config values use this format:

```text
<type>:<source>
```

Supported types:

| Type | Example | Meaning |
|---|---|---|
| `envs` | `envs:UNSEALING_PASSPHRASE` | Read bytes from an environment variable |
| `file` | `file:/run/secrets/seed.seal` | Read bytes from a file |

Example environment:

```bash
export UNSEALING_PASSPHRASE='strong passphrase'
export LITEHSM_APP_KEYMANAGER_SEED_PASSPHRASE='envs:UNSEALING_PASSPHRASE'
export LITEHSM_APP_KEYMANAGER_SEED_SEALING='file:/etc/litehsm/seed.seal'

export LITEHSM_COMMON_DB_HOST='localhost'
export LITEHSM_COMMON_DB_PORT='5432'
export LITEHSM_COMMON_DB_DATABASE='litehsm'
export LITEHSM_COMMON_DB_USER='litehsm'
export LITEHSM_COMMON_DB_PASSWORD='litehsm'
export LITEHSM_COMMON_GRPC_PORT='50051'
```

---

## Seed Initialization

Use `litehsmctl init` to generate a random 32-byte seed and write it as a sealed envelope.

Build the CLI:

```bash
go build -o litehsmctl ./cmd/litehsmctl
```

Initialize from an environment variable passphrase:

```bash
export HSM_PASSPHRASE='strong passphrase'
./litehsmctl init \
  -from env \
  -input HSM_PASSPHRASE \
  -output ./seed.seal
```

Initialize from a passphrase file:

```bash
./litehsmctl init \
  -from file \
  -input ./passphrase.txt \
  -output ./seed.seal
```

Initialize from stdin:

```bash
./litehsmctl init \
  -from stdin \
  -output ./seed.seal
```

The output file is written with `0600` permissions.

---

## gRPC API

The service registers `services.litehsm.v1.KeyManagerService` from [`github.com/LiquidCats/paw/protos`](https://github.com/LiquidCats/paw/protos).

Implemented RPCs:

| RPC | Purpose |
|---|---|
| `CreateKey` | Creates a disabled key metadata record with a generated derivation path |
| `SetKeyExpiration` | Updates a key expiration timestamp |
| `EnableKey` | Sets key status to `enabled` |
| `DisableKey` | Sets key status to `disabled` |
| `DeleteKey` | Sets key status to `deleted` |

Supported domain values:

| Field | Supported Values |
|---|---|
| Curve | `secp256k1` |
| Algorithm | `ecdsa` |
| Key status | `enabled`, `disabled`, `deleted` |

Example `grpcurl` call:

```bash
grpcurl -plaintext \
  -d '{"alias":"main-key","curve":"CURVE_SECP256K1","algorithm":"ALGORITHM_ECDSA"}' \
  localhost:50051 \
  services.litehsm.v1.KeyManagerService/CreateKey
```

---

## Database

`litehsm` uses PostgreSQL and runs embedded migrations from `internal/adapter/postgresql/database/migrations` at startup.

Current schema areas:

- `keys`: key metadata, alias, curve, algorithm, derivation path, status, timestamps, optional expiration.
- `chains`: BIP-44 coin metadata seeded from migrations.
- `event_log`: JSON event payload storage.

SQL is generated with `sqlc` from:

```text
internal/adapter/postgresql/database/queries
internal/adapter/postgresql/database/migrations
```

Regenerate generated database code with:

```bash
make db-gen
```

---

## Running Locally

Start PostgreSQL and create a database/user matching your environment variables, then initialize a seed envelope.

```bash
go build -o litehsm ./cmd/litehsm
go build -o litehsmctl ./cmd/litehsmctl

export HSM_PASSPHRASE='strong passphrase'
./litehsmctl init -from env -input HSM_PASSPHRASE -output ./seed.seal

export LITEHSM_APP_KEYMANAGER_SEED_PASSPHRASE='envs:HSM_PASSPHRASE'
export LITEHSM_APP_KEYMANAGER_SEED_SEALING='file:./seed.seal'
export LITEHSM_COMMON_DB_HOST='localhost'
export LITEHSM_COMMON_DB_PORT='5432'
export LITEHSM_COMMON_DB_DATABASE='litehsm'
export LITEHSM_COMMON_DB_USER='litehsm'
export LITEHSM_COMMON_DB_PASSWORD='litehsm'

./litehsm
```

---

## Docker

The Dockerfile builds `cmd/litehsm` with Go 1.26.4 and runs it on a distroless Debian 12 image.

```bash
docker build -t litehsm:latest .
```

Example run:

```bash
docker run --rm \
  -p 50051:50051 \
  -v "$PWD/seed.seal:/seed.seal:ro" \
  -e HSM_PASSPHRASE='strong passphrase' \
  -e LITEHSM_APP_KEYMANAGER_SEED_PASSPHRASE='envs:HSM_PASSPHRASE' \
  -e LITEHSM_APP_KEYMANAGER_SEED_SEALING='file:/seed.seal' \
  -e LITEHSM_COMMON_DB_HOST='host.docker.internal' \
  -e LITEHSM_COMMON_DB_PORT='5432' \
  -e LITEHSM_COMMON_DB_DATABASE='litehsm' \
  -e LITEHSM_COMMON_DB_USER='litehsm' \
  -e LITEHSM_COMMON_DB_PASSWORD='litehsm' \
  litehsm:latest
```

The image exposes port `8080` in the current Dockerfile, while the application defaults to gRPC port `50051`. Set `LITEHSM_COMMON_GRPC_PORT` explicitly if you run the service on a different port.

---

## Development

### Prerequisites

- Go 1.26.4+
- Docker, for mock generation, linting, sqlc generation, and containerized test/benchmark targets
- PostgreSQL, for running the service locally
- Access to private Go modules under `github.com/LiquidCats`, when required by your environment

### Test

```bash
go test ./...
```

Containerized test target:

```bash
make run-tests
```

### Benchmark

```bash
go test -bench=. ./...
```

Containerized benchmark target:

```bash
make run-bench
```

### Mocks

Mocks are generated with `vektra/mockery` v3 into `test/mocks/litehsm`.

```bash
make mock
```

### Lint

```bash
make lint
make lint-fix
```

### SQLC

```bash
make db-gen
```

### Build

```bash
go build -o litehsm ./cmd/litehsm
go build -o litehsmctl ./cmd/litehsmctl
```
