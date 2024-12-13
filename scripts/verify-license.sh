#!/usr/bin/env bash

set -ex

CONFIG_PATH=${CONFIG_PATH:-$(dirname "${BASH_SOURCE[0]}")/..}

if license-lint -config ${CONFIG_PATH}/license-lint.yml; then
    echo "✅ License lint succeeded"
else
    echo # print one empty line, separate from warning messages.
    echo '❌ Please review the above error messages.'
    exit 1
fi
