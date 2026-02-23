#!/bin/sh
# install.sh — Install sonar via go install
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/ar1o/sonar/main/scripts/install.sh | sh
#
# Requires Go 1.24+ to be installed.

set -eu

main() {
    if ! command -v go >/dev/null 2>&1; then
        err "go is required but not found in PATH"
        err "install Go from https://go.dev/dl/ and try again"
        exit 1
    fi

    info "installing sonar..."
    go install github.com/ar1o/sonar/cmd/sonar@latest

    GOBIN="$(go env GOPATH)/bin"
    info "installed sonar to ${GOBIN}/sonar"

    # PATH guidance
    case ":${PATH}:" in
        *":${GOBIN}:"*) ;;
        *)
            warn "${GOBIN} is not in your PATH"
            info "add it by appending this to your shell profile:"
            info "  export PATH=\"${GOBIN}:\$PATH\""
            ;;
    esac
}

info() {
    printf '[sonar] %s\n' "$1" >&2
}

warn() {
    printf '[sonar] warning: %s\n' "$1" >&2
}

err() {
    printf '[sonar] error: %s\n' "$1" >&2
}

main "$@"
