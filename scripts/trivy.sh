#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# ignore VULNEEABILITY CVE-2022-1996 it will fix at k8s.io/api next release
# ignore unfixed  VULNEEABILITY

TRIVY_DB_REPOSITORY=${TRIVY_DB_REPOSITORY:-ghcr.io/aquasecurity/trivy-db}

trivy fs --scanners secret --secret-config ./.trivycert.yaml --exit-code 1 ./

# The parameters that this shell receives look like this ：
# HIGH,CRITICAL release-ci.daocloud.io/mspider/mspider:v0.8.3-47-gd3ac6536  release-ci.daocloud.io/mspider/mspider-api-server:v0.8.3-47-gd3ac6536
# so need use firtParameter parameter to skip first Parameter HIGH,CRITICAL than trivy images
firtParameter=1
for i in "$@"; do
    if (($firtParameter == 1)); then
        ((firtParameter = $firtParameter + 1))
    else
        trivy image --skip-dirs istio.io/istio --ignore-unfixed --db-repository=${TRIVY_DB_REPOSITORY} --exit-code 1 --severity $1 $i
    fi
done
