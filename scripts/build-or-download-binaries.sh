#!/usr/bin/env bash

set -ex

CUR_DIR=$(
    cd -- "$(dirname "$0")" >/dev/null 2>&1
    pwd -P
)

PLATFORMS=${PLATFORMS:-linux/amd64}
APP=${APP:-knoway-gateway}

for p in $(echo ${PLATFORMS} | tr "," " "); do
    GOOS=$(echo ${p} | cut -d "/" -f 1)
    GOARCH=$(echo ${p} | cut -d "/" -f 2)
    dist=${CUR_DIR}/../out/$p/
    mkdir -p ${dist}
    echo "building ${APP} for ${GOOS}/${GOARCH}"
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-s -w" -o ${CUR_DIR}/../out/$p/${APP} ${CUR_DIR}/../cmd
done
