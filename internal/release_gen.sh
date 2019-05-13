#!/usr/bin/env bash

set -euo pipefail

DIR="$(cd "$(dirname "${0}")/.." && pwd)"
cd "${DIR}"

goos() {
  case "${1}" in
    Darwin) echo darwin ;;
    Linux) echo linux ;;
    *) return 1 ;;
  esac
}

goarch() {
  case "${1}" in
    x86_64) echo amd64 ;;
    *) return 1 ;;
  esac
}

sha256() {
  if ! type sha256sum >/dev/null 2>/dev/null; then
    if ! type shasum >/dev/null 2>/dev/null; then
      echo "sha256sum and shasum are not installed" >&2
      return 1
    else
      shasum -a 256 "$@"
    fi
  else
    sha256sum "$@"
  fi
}

RELEASE_DIR="release"

rm -rf "${RELEASE_DIR}"
for os in Darwin Linux; do
  for arch in x86_64; do
    for binary_name in protoc-gen-twirp protoc-gen-twirp_python; do
      BINARY="${RELEASE_DIR}/${binary_name}-${os}-${arch}"
      CGO_ENABLED=0 GOOS=$(goos "${os}") GOARCH=$(goarch "${arch}") \
        go build -a -installsuffix cgo -o "${BINARY}" \
        $(find "${binary_name}" -name '*.go')
      sha256 "${BINARY}" > "${BINARY}.sha256sum"
      sha256 -c "${BINARY}.sha256sum"
    done
  done
done
