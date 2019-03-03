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

mole_syntax
