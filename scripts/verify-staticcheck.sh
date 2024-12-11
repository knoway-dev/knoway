#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
GOLANGCI_LINT_PKG="github.com/golangci/golangci-lint/cmd/golangci-lint"
GOLANGCI_LINT_VER="v1.53.3"

cd "${REPO_ROOT}"
source "scripts/util.sh"

command golangci-lint &>/dev/null || util::install_tools ${GOLANGCI_LINT_PKG} ${GOLANGCI_LINT_VER}

golangci-lint --version

if golangci-lint run -v --timeout=5m; then
    echo '✅ Congratulations! All Go source files have passed staticcheck.'
else
    echo '❌ Staticcheck failed. Please review the warnings above.'
    echo '💡 Tip: If these warnings are unclear, you can file an issue for help.'
    exit 1
fi
