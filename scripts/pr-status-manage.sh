#!/usr/bin/env bash

# 这个脚本不止做 Cherry Pick，后期可能还会集成其它功能，比如 Coverage 检测？
# 所以名字叫 PR Status Manage

set -ex

export GITLAB_HOST=${GITLAB_HOST:-https://gitlab.daocloud.cn}

PURE_HOST=${GITLAB_HOST//https:\/\//}

APIREPO=${CI_PROJECT_PATH//\//%2F}

glab auth login -t ${GITLAB_CI_TOKEN} -h ${PURE_HOST}
glab auth status

git fetch

cherrypickto() {
    target_branch=$1
    branch=cherrypick-${CI_MERGE_REQUEST_IID}-to-${target_branch}
    git branch -D $branch || true # force delete local branch if exists
    git checkout ${target_branch} # checkout to target branch
    git checkout -b $branch       # checkout to a new branch

    # get current pr commits and reserve by date
    commits=$(glab api projects/${APIREPO}/merge_requests/${CI_MERGE_REQUEST_IID}/commits | jq '.[].id' -r | sed '1!G;h;$!d')
    echo '```' >/tmp/cherry-pick.log
    echo "Auto cherry-pick !${CI_MERGE_REQUEST_IID} to ${target_branch} failed!" >>/tmp/cherry-pick.log
    git config user.name "Auto Cherry-pick Bot"
    git config user.email "cherry-pick-bot@daocloud.io"
    for commit in ${commits}; do
        if ! git cherry-pick ${commit} --allow-empty &>>/tmp/cherry-pick.log; then
            echo "cherry-pick ${commit} failed"
            echo '```' >>/tmp/cherry-pick.log
            cat /tmp/cherry-pick.log
            # check if already exists failed issue
            if glab issue list --in title --search "[manually cherry-pick required] !${CI_MERGE_REQUEST_IID}" | grep "!${CI_MERGE_REQUEST_IID}"; then
                echo "Issue already exists, skip"
                return
            fi
            author=$(glab mr view ${CI_MERGE_REQUEST_IID} | awk '/^author:/ {print $2}')
            # create an issue if cherry-pick failed.
            glab issue create \
                --title "[manually cherry-pick required] !${CI_MERGE_REQUEST_IID} Auto cherry-pick to ${target_branch} error" \
                --description "$(cat /tmp/cherry-pick.log)" \
                --assignee ${author}
            return
        fi
    done
    title="Auto cherry-pick !${CI_MERGE_REQUEST_IID} to ${target_branch}"
    mr_state=$(glab mr view ${CI_MERGE_REQUEST_IID} | awk '/^state:/ {print $2}')
    if [[ "${mr_state}" != "merged" ]]; then
        # if mr is not merged, mark new mr as draft to avoid merge it by mistake
        title="Draft: ${title}"
    fi
    git remote set-url origin https://gitlab-ci-token:${GITLAB_CI_TOKEN}@${PURE_HOST}/${CI_PROJECT_PATH}.git
    git push origin ${branch} -f
    if ! glab mr list --source-branch=${branch} --target-branch=${target_branch} | grep "Auto cherry-pick"; then
        res=$(glab mr create --no-editor \
            --remove-source-branch \
            --source-branch ${branch} \
            --target-branch ${target_branch} \
            --title "${title}" \
            --label auto-cherry-picked \
            --description "Auto cherry-pick from !${CI_MERGE_REQUEST_IID}" &>/dev/stdout)
        glab mr note ${CI_MERGE_REQUEST_IID} -m "### Auto cherry-picked!<br>${res}"
    else
        echo "MR already exists, skip"
    fi
}

cherrypick() {
    if [ -z "${CI_MERGE_REQUEST_LABELS}" ]; then
        echo "No cherry-pick labels found."
        return
    fi
    for label in $(echo "${CI_MERGE_REQUEST_LABELS}" | tr ',' '\n'); do
        if [[ ${label} == cherry-pick-* ]]; then
            target=${label//cherry-pick-/}
            if ! git rev-list origin/${target} >/dev/null; then
                echo "target branch ${target} not exists" >/dev/stderr
                exit 1
            fi
            cherrypickto ${target}
        fi
    done
}

cherrypick
