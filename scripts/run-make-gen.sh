#!/usr/bin/env bash

set -ex

CUR_DIR=$(
    cd -- "$(dirname "$0")" >/dev/null 2>&1
    pwd -P
)

git fetch origin ${CI_BUILD_REF_NAME}
git checkout ${CI_BUILD_REF_NAME}

cd ${CUR_DIR}/..
make gen

git rev-parse --abbrev-ref HEAD
git status

if [ "$(git status --porcelain | wc -l)" -eq "0" ]; then
    echo "${DIFFPROTO} up to date."
    exit 0
fi

if [ -n "${GITLAB_TOKEN}" ]; then
    git remote set-url origin https://gitlab-ci-token:${GITLAB_TOKEN}@gitlab.daocloud.cn/${CI_PROJECT_PATH}.git
fi

if ! git config user.name; then
    git config user.name "Auto Gen Bot"
    git config user.email "auto-gen-bot@daocloud.io"
fi

git add .

git commit -m "Auto run gen code"

git push origin ${CI_BUILD_REF_NAME}
