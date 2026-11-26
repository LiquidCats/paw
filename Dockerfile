# syntax=docker/dockerfile:1@sha256:572cebbc2b15e6f84c077742952674889cb63cd5e9b80abdffcbe587b6f19183

# Builds a Go service with Nix. The Go toolchain and every dependency come from
# the nixpkgs revision pinned in flake.lock (see flake.nix).
#
# SOURCE is the service to build: a directory name in ./cmd, e.g. rater. It's required.
#   docker build --build-arg SOURCE=rater .
#   SOURCE=rater docker build --build-arg SOURCE .   # value taken from env
#
# For identical image digests, not only identical binaries:
#   SOURCE_DATE_EPOCH=$(git log -1 --format=%ct) docker buildx build \
#     --build-arg SOURCE=rater --build-arg SOURCE_DATE_EPOCH --provenance=false .
#
# Without Docker, on any Nix machine: nix build .#rater

ARG NIX_IMAGE=nixos/nix:2.36.0@sha256:746280551ddc1c72177d1a597c05eacac86dfaa18d5d649bf86411278bcd5470
ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

FROM ${NIX_IMAGE} AS build

# filter-syscalls=false: Nix's seccomp filter can't load under QEMU (cross-arch buildx)
RUN printf 'experimental-features = nix-command flakes\nfilter-syscalls = false\n' >> /etc/nix/nix.conf
WORKDIR /src

# fetch the pinned nixpkgs and Go toolchain; cached until flake.lock changes
COPY flake.nix flake.lock ./
RUN nix build --no-link --inputs-from . nixpkgs#go_1_27

COPY . .

ARG SOURCE
ARG SOURCE_DATE_EPOCH
RUN : "${SOURCE:?build arg SOURCE is required: a directory name in ./cmd, e.g. rater}" && \
    nix build --print-build-logs ".#${SOURCE}" && \
    install -D -m 0755 "result/bin/${SOURCE}" /out/app && \
    touch -d "@${SOURCE_DATE_EPOCH:-0}" /out/app

FROM ${RUNTIME_IMAGE}
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
