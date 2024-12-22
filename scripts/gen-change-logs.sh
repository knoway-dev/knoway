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

CUR_VERSION=${CUR_VERSION:-v0.0.0}

OUTFILE=${1:-${CUR_DIR}/../changes/CHANGELOG-${CUR_VERSION}.md}

mkdir -p $(dirname ${OUTFILE})

getallmrs() {
    git fetch origin --tags ${PRE_VERSION} &>/dev/null
    git log ${CI_BUILD_REF_NAME} ^${PRE_VERSION} | grep -E '\(![^\)]+\)$' | sed -r 's/.*\(\!(.*)\)$/\1/g' | uniq | sort
}

features=""
bugs=""

for mr in $(getallmrs); do
    cont=$(glab mr view ${mr})
    title=$(echo "${cont}" | grep -E '^title:' | sed 's/title:\t//g')
    author=$(echo "${cont}" | grep -E '^author:' | sed 's/author:\t//g')
    labels=$(echo "${cont}" | grep -E '^labels:' | sed 's/labels:\t//g')
    if echo "${labels}" | grep -E 'kind/feature' &>/dev/null; then
        echo "mr ${mr} is a feature"
        features+="- ${title}(!${mr}) by @${author}
"
    else
        echo "mr ${mr} is not a feature"
        bugs+="- ${title}(!${mr}) by @${author}
"
    fi
done

echo "
# ${CUR_VERSION} Change logs

## Change since ${PRE_VERSION}

### Changes by Kind

#### Bug

${bugs}

#### Feature

${features}

" >${OUTFILE}
