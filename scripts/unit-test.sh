#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

set -x
PATH2TEST=(./pkg/... ./internal/...)
tmpDir=$(mktemp -d)
mergeF="${tmpDir}/merge.out"
rm -f ${mergeF}
for ((i = 0; i < ${#PATH2TEST[@]}; i++)); do
    ls $tmpDir
    cov_file="${tmpDir}/$i.cover"
    GOMAXPROCS=8 go test --race --v -covermode=atomic -coverpkg=${PATH2TEST[i]} -coverprofile=${cov_file} ${PATH2TEST[i]} # $(go list ${PATH2TEST[i]})
    cat $cov_file | grep -v mode: >>${mergeF} || echo no coverage found
done
#merge them
header=$(head -n1 "${tmpDir}/0.cover")
echo "${header}" >coverage.out
cat ${mergeF} >>coverage.out
go tool cover -func=coverage.out
rm -rf coverage.out ${tmpDir} ${mergeF}
