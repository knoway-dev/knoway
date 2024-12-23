#!/usr/bin/env bash

set -ex

CUR_DIR=$(
    cd -- "$(dirname "$0")" >/dev/null 2>&1
    pwd -P
)

export GITLAB_HOST=${GITLAB_HOST:-https://gitlab.daocloud.cn}

PURE_HOST=${GITLAB_HOST//https:\/\//}

glab auth login -t ${GITLAB_CI_TOKEN} -h ${PURE_HOST}
glab auth status

minor_version=$(grep "MINOR_VERSION ?=" ${CUR_DIR}/../Makefile | sed -r 's/MINOR_VERSION \?= (.*)/\1/g' | xargs)

if [ -n "${NEXT_VERSION}" ] && ! echo ${NEXT_VERSION} | grep -E "^v\d+\.\d+$"; then
    echo "Invalid NEXT_VERSION: ${NEXT_VERSION}, require running on v*.* branch"
    exit 1
else
    major_number=$(echo ${minor_version} | awk -F. '{print $1}')
    minor_number=$(echo ${minor_version} | awk -F. '{print $2}')
    NEXT_VERSION=v${major_number}.$((${minor_number} + 1))
fi

if [ -n "${GITLAB_CI_TOKEN}" ]; then
    git remote set-url origin https://gitlab-ci-token:${GITLAB_CI_TOKEN}@gitlab.daocloud.cn/ndx/ai/knoway.git
fi

if git ls-remote --exit-code origin release-${minor_version} &>/dev/null; then
    echo "release-${minor_version} branch already exists"
    exit 1
fi

if ! git config user.name; then
    git config user.name "Auto Release Bot"
    git config user.email "knoway-auto-release@daocloud.io"
fi

git checkout -b release-${minor_version}

# change version
if [ "$(uname)" = "Darwin" ]; then
    sed -i "" "s/MINOR_VERSION ?=.*/MINOR_VERSION ?= ${NEXT_VERSION}/g" ${CUR_DIR}/../Makefile
else
    sed -i "s/MINOR_VERSION ?=.*/MINOR_VERSION ?= ${NEXT_VERSION}/g" ${CUR_DIR}/../Makefile
fi

# push release branch
git push origin release-${minor_version}

# create label
glab label create --color="#ed9121" -n cherry-pick-release-${minor_version}

git checkout ${CI_BUILD_REF_NAME}

git add ${CUR_DIR}/../Makefile
git commit -m "Code freeze and bump MINOR_VERSION to ${NEXT_VERSION}"

# push origin branch
git push origin ${CI_BUILD_REF_NAME}
