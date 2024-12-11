#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

DIFFPROTO="${SCRIPT_ROOT}"
make gen
if [ "$(git status --porcelain | wc -l)" -eq "0" ]; then
    echo "${DIFFPROTO} up to date."
else
    echo "${DIFFPROTO} is out of date. Please run make gen to update codes for the proto files."
    echo "Diff files:"
    git status --porcelain
    git diff
    exit 1
fi
