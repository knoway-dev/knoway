#!/usr/bin/env bash

set -x
set -o errexit
set -o nounset
set -o pipefail

temp=$(mktemp)
new_version() {
    echo '{{- if semverCompare ">=1.23.0-0" .Capabilities.KubeVersion.GitVersion }}' >$2
    cat $1 >>$2
    echo '{{- else }}' >>$2
}
old_version() {
    cat $1 >>$2
    echo '{{- end }}' >>$2
}

if [[ "" == $(cat $1 | yq '.. | select(has("x-kubernetes-validations"))') ]]; then
    echo "no x-kubernetes-validations found, skip"
    cp $1 $2
    exit
fi

f=$(basename $1)
new_version $1 $2/${f}

cat $1 | yq 'del(.. | select(has("x-kubernetes-validations")).x-kubernetes-validations)' >$temp

old_version $temp $2/${f}

rm -f ${temp}
