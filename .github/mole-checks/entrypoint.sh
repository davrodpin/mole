#!/bin/sh -l

GO_INSTALL_DIR=$HOME

[ -z "$GOROOT" ] && export GOROOT=${GO_INSTALL_DIR}/go
[ -z "$GOPATH" ] && export GOPATH=$HOME/go_workspace

MOLE_SRC_PATH=${GOPATH}/src/github.com/${GITHUB_REPOSITORY}
GO=$GOROOT/bin/go
COV_PROFILE=${GITHUB_WORKSPACE}/coverage.out
JSON='{ "message": "{{MESSAGE}}", "committer": { "name": "Mole Bot", "email": "davrodpin+molebot@gmail.com" }, "content": "{{CONTENT}}" }'

log() {
  level="$1"
  message="$2"

  [ -z "$level" ] || [ -z "$message" ] && return 1

  printf "`date +%Y-%m-%dT%H:%M:%S%z`\t%s\t%s\n" "${level}" "${message}"

  return 0
}


mole_create_wksp() {
  log "info" "Creating Go workspace at ${GOPATH}"

  [ ! -d "${GOPATH}" ] && {
    mkdir -p $GOPATH/{src,bin,pkg} && \
      mkdir -p ${MOLE_SRC_PATH} && \
      cp -a $GITHUB_WORKSPACE/* ${MOLE_SRC_PATH} || return 1
  }

  return 0
}

go_install() {
  [ ! -f "$GO" ] && {
    cd ${GO_INSTALL_DIR} && \
      log "info" "downloading https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz" && \
      curl -O https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz && \
      tar -C $HOME -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
      log "info" "go ${GO_VERSION} installed with success" || return 1
  }

  return 0
}

mole_syntax() {
  go_install && mole_create_wksp  || return 1

  log "info" "looking for code formatting issues on mole..." || return 1
  fmt=`$GO fmt github.com/${GITHUB_REPOSITORY}/... | sed 's/\n/ /p'`
  retcode=$?

  [ -n "$fmt" ] && {
    log "error" "the following files do not follow the Go syntax conventions: ${fmt}"
    return 1
  }

  return $retcode
}

mole_test() {
  go_install && \
    mole_create_wksp && \
    log "info" "running mole's tests..." &&  \
    $GO test github.com/${GITHUB_REPOSITORY}/... -v -race -coverprofile=${COV_PROFILE} -covermode=atomic || return 1

  return 0
}

commit() {
  path="$1"
  message="$2"
  content="$3"

  [ -z "$path" ] || [ -z "$message" ] || [ -z "$content" ] && {
    log "info" "could not publish new report"
    return 1
  }

	payload=`printf "$JSON\n" | sed "s:{{MESSAGE}}:${message}:"`
	payload=`printf "$payload\n" | sed "s:{{CONTENT}}:${content}:"`

  uri="/repos/davrodpin/mole/contents/${path}"
  url="https://api.github.com${uri}"

  log "info" "Adding new coverage report to mole source code repo..."
  response=`curl -v -H "Authorization: token ${GITHUB_TOKEN}" -X PUT "${url}" -d "${payload}" 2>&1`
  resp_code=`printf "%s\n" "${response}" | sed -n 's/< HTTP\/\([.0-9]\{1,\}\) \([0-9]\{3\}\)\(.\{0,\}\)/\2/p' | tail -n1`

  log "debug" "PUT ${uri}"
  log "debug" "response from api.github.com: ${resp_code}"

  [ "${resp_code}" != "201" ] && {
    log "error" "error while publishing the new report to mole source code repo..."
    printf "\n\n%s\n\n" "${response}"
    return 1
  }

  return 0
}

mole_report() {
  [ ! -f "$COV_PROFILE" ] && {
    log "error" "Coverage Profile not found ${COV_PROFILE}"
    return 1
  }

  go_install || return 1

  prev_commit_id=`jq '.before' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`
  commit_id=`jq '.after' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`
  report_name=mole-cov-`date +%s`.html
  report_path=${GITHUB_WORKSPACE}/${report_name}

  $GO tool cover -html=${COV_PROFILE} -o ${report_path} || {
    log "error" "error generating coverage report"
    return 1
  }

  path="docs/cov/${commit_id}/${report_name}"
	message="Add new coverage report"
  content=`base64 -w0 $report_path`

  commit "$path" "$message" "$content"
  retcode=$?

  [ $retcode -eq 0 ] && {
    report_url="http://htmlpreview.github.io/?https://github.com/${GITHUB_REPOSITORY}/blob/master/${path}"
    log "info" "coverage report available on ${report_url}"
  }

  return $retcode
}

case "$1" in
  "syntax")
    mole_syntax
    ;;
  "test")
    mole_test
    ;;
  "report")
    mole_report
    ;;
  *)
    sh -c "$*"
    ;;
esac
