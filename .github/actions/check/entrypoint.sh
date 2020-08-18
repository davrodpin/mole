#!/bin/sh -l

export GOPATH=/go

GO="/usr/local/go/bin/go"
GOMPLATE="/usr/local/bin/gomplate"
MOLE_SRC_PATH=${GOPATH}/src/github.com/${GITHUB_REPOSITORY}
COV_PROFILE=${GITHUB_WORKSPACE}/coverage.txt
COV_REPORT=${GITHUB_WORKSPACE}/mole-coverage.html
COV_DIFF_DATA=${GITHUB_WORKSPACE}/cov-diff.json
COV_DIFF_HTML_REPORT_TPL=/cov-diff.html.tpl
COV_DIFF_TXT_REPORT_TPL=/cov-diff.txt.tpl
COV_DIFF_REPORT=${GITHUB_WORKSPACE}/mole-diff-coverage.html
COV_DIFF_DATA_TPL='{ "Title": "{{TITLE}}", "Previous_Commit": "{{PREV_COMMIT}}", "Current_Commit": "{{CURR_COMMIT}}", "Created_At": "{{CREATED_AT}}", "Files": [{{ROWS}}] }'
COV_DIFF_DATA_ROW_TPL='{ "File": "{{FILE}}", "Previous_Coverage": {{PREV}}, "Current_Coverage": {{CUR}}, "Diff": {{DIFF}} }'

log() {
  level="$1"
  message="$2"

  [ -z "$level" ] || [ -z "$message" ] && return 1

  printf "`date +%Y-%m-%dT%H:%M:%S%z`\t%s\t%s\n" "${level}" "${message}"

  return 0
}

mole_wksp() {
  log "info" "Creating Go workspace at ${GOPATH}"

  mkdir -p ${MOLE_SRC_PATH} && \
    cp -a $GITHUB_WORKSPACE/* ${MOLE_SRC_PATH} || return 1

  return 0
}

download_report() {
  commit="$1"

  [ -z "$commit" ] && return 1

  resp=`curl --silent --show-error -X POST https://content.dropboxapi.com/2/files/download \
    --header "Authorization: Bearer ${DROPBOX_TOKEN}" \
    --header "Dropbox-API-Arg: {\"path\": \"/reports/${commit}/mole-coverage.html\"}"`

  error=`printf "%s\n" "$resp" | jq 'select(.error != null) | .error' 2> /dev/null`
  [ -n "$error" ] && {
    log "debug" "${resp}"
    return 1
  }

  printf "%s\n" "${resp}"

  return 0
}

mole_test() {
  prev_commit_id=`jq '.before' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`
  commit_id=`jq '.after' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`

  mole_wksp || return 1

  ## TEST

  log "info" "running mole's tests and generating coverage profile for ${commit_id}"
  $GO test github.com/${GITHUB_REPOSITORY}/... -v -race -coverprofile=${COV_PROFILE} -covermode=atomic || return 1

  $GO tool cover -html=${COV_PROFILE} -o ${COV_REPORT} || {
    log "error" "error generating coverage report"
    return 1
  }

  log "info" "looking for code formatting issues on mole@${commit_id}" || return 1
  fmt=`$GO fmt github.com/${GITHUB_REPOSITORY}/... | sed 's/\n/ /p'`
  retcode=$?

  if [ -n "$fmt" ]
  then
    log "error" "the following files do not follow the Go formatting convention: ${fmt}"
    return ${retcode}
  else
    log "info" "all source code files are following the formatting Go convention"
  fi

  #TODO publish list of files failing on go-fmt

  #TODO Use $GITHUB_REF to post comment back to PR
  #TODO warning if code coverage decreases

  return 0
}

mole_test
#TODO return 1 if error and 78 if check fails

