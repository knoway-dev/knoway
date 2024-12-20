#!/usr/bin/env bash

set -ex

CUR_DIR=$(
    cd -- "$(dirname "$0")" >/dev/null 2>&1
    pwd -P
)

MINOR_VERSION=$(echo ${CI_BUILD_REF_NAME} | sed -r 's/^release-(.*)$/\1/g')

if [[ ${CI_BUILD_REF_NAME} == ${MINOR_VERSION} || ${MINOR_VERSION} == "" ]]; then
    echo "Invalid branch name: ${CI_BUILD_REF_NAME}, require running on release-.* branch"
    exit 1
fi

if [ -z "${PRE_VERSION}" ]; then
    echo you must specify PRE_VERSION var >>/dev/stderr
    exit 1
fi

if [ -z "${PATCH_VERSION}" ]; then
    echo you must specify NEXT_VERSION var >>/dev/stderr
    exit 1
fi

CUR_VERSION=v${MINOR_VERSION}.${PATCH_VERSION}

SHORT_VERSION=v${MINOR_VERSION}

echo "VERSION is ${CUR_VERSION}"
echo "SHORT_VERSION is ${SHORT_VERSION}"

if [ "${PRE_VERSION}" = "${CUR_VERSION}" ]; then
    echo PRE_VERSION should not be same as knoway in current version >>/dev/stderr
    exit 1
fi

git fetch

if ! git rev-list ${PRE_VERSION} >/dev/null; then
    echo "${PRE_VERSION} tag not exists" >/dev/stderr
    exit 1
fi

if [ -n "${CI_BUILD_REF_NAME}" ]; then
    git checkout ${CI_BUILD_REF_NAME}
fi

# todo release notes
#cd ${CUR_DIR}/../tools/gen-release-notes
#mkdir -p ${CUR_DIR}/../changes/${SHORT_VERSION}
#go run . --oldRelease ${PRE_VERSION} --newRelease ${CUR_VERSION} --notes ${CUR_DIR}/../ --outDir ${CUR_DIR}/../changes/${SHORT_VERSION}

CUR_VERSION=${CUR_VERSION} bash ${CUR_DIR}/gen-change-logs.sh ${CUR_DIR}/../changes/${SHORT_VERSION}/CHANGELOG-${CUR_VERSION}.md

cd ${CUR_DIR}/..

if ! git config user.name; then
    git config user.name "Auto Release Bot"
    git config user.email "knoway-auto-release@daocloud.io"
fi

# we no need to sync api repo any more
# sh ${CUR_DIR}/sync-api-repo.sh ${CUR_VERSION}

cd ${CUR_DIR}/..

git add .

git commit -m "Release ${CUR_VERSION} and add release notes"

cat ${CUR_DIR}/../changes/${SHORT_VERSION}/CHANGELOG-${CUR_VERSION}.md | git tag -a ${CUR_VERSION} -F-

if [ -n "${GITLAB_CI_TOKEN}" ]; then
    git remote set-url origin https://gitlab-ci-token:${GITLAB_CI_TOKEN}@gitlab.daocloud.cn/ndx/ai/knoway.git
fi

# push to release branch
if [ -z "${CI_BUILD_REF_NAME}" ]; then
    git push origin $(git rev-parse --abbrev-ref HEAD)
else
    git push origin ${CI_BUILD_REF_NAME}
fi

COMMIT=$(git rev-parse HEAD)

# Push release notes to main branch also
git checkout main
git cherry-pick ${COMMIT}
git push origin main

# push tag
git push origin ${CUR_VERSION}

curl -s -v \
    -H "PRIVATE-TOKEN: ${GITLAB_CI_TOKEN}" \
    -H 'Content-Type: application/json' \
    'https://gitlab.daocloud.cn/api/v4/projects/ndx%2Fai/knoway/releases' \
    -X POST \
    -d "$(echo '{}' | jq \
        --arg name "Release ${CUR_VERSION}" \
        --arg tag_name "${CUR_VERSION}" \
        --arg description "$(cat ${CUR_DIR}/../changes/${SHORT_VERSION}/CHANGELOG-${CUR_VERSION}.md)" \
        '.name = $name | .tag_name = $tag_name | .description = $description')"
