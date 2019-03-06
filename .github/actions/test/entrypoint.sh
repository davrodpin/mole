#!/bin/sh -l

GO_INSTALL_DIR=$HOME

export GOROOT=${GO_INSTALL_DIR}/go
export GOPATH=$HOME/go_workspace

MOLE_SRC_PATH=${GOPATH}/src/github.com/${GITHUB_REPOSITORY}
GO=$GOROOT/bin/go
COV_PROFILE=${GITHUB_WORKSPACE}/coverage.out
COV_REPORT=${GITHUB_WORKSPACE}/mole-coverage.html
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
      curl --silent --show-error -O https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz && \
      tar -C $HOME -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
      log "info" "go ${GO_VERSION} installed with success" || return 1
  }

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

cov_diff() {
  prev="$1"

  [ ! -f "$COV_REPORT" ] && {
    log "error" "coverage diff can't be computed: report file is missing: ${COV_REPORT}"
    return 1
  }

  prev_report=`download_report "${prev}"`
  [ $? -ne 0 ] && {
    log "warn" "coverage diff can't be computed: report could not be donwloaded for ${prev}"
    printf "%s\n" "${prev_report}"
    return 2
  }

  curr_stats=`cat ${COV_REPORT} | grep "<option value=" | sed -n 's/[[:blank:]]\{0,\}<option value="file[0-9]\{1,\}">\(.\{1,\}\) (\([0-9.]\{1,\}\)%)<\/option>/\1,\2/p'`
  prev_stats=`echo "${prev_report}" | grep "<option value=" | sed -n 's/[[:blank:]]\{0,\}<option value="file[0-9]\{1,\}">\(.\{1,\}\) (\([0-9.]\{1,\}\)%)<\/option>/\1,\2/p'`

  [ -z "$curr_stats" ] || [ -z "$prev_stats" ] && {
    log "error" "could not extract the code coverage numbers from ${COV_REPORT} and/or ${prev}"
    return 1
  }

  for stats1 in `printf "%s\n" "$curr_stats"`
  do
    mod1=`printf "%s\n" "$stats1" | awk -F, '{print $1}'`
    cov1=`printf "%s\n" "$stats1" | awk -F, '{print $2}'`
    diff=0

    for stats2 in `printf "%s\n" "$prev_stats"`
    do
      mod2=`printf "%s\n" "$stats2" | awk -F, '{print $1}'`
      cov2=`printf "%s\n" "$stats2" | awk -F, '{print $2}'`

      [ "$mod1" = "$mod2" ] && {
        diff=`printf "%s - %s\n" "$cov1" "$cov2" | bc`
        break
      }
    done

    printf "[mod=%s, cov=%s]\n" "$mod1" "$diff"
  done

  return 0
}

mole_test() {
  prev_commit_id=`jq '.before' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`
  commit_id=`jq '.after' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`

  go_install && \
    mole_create_wksp && \
    log "info" "running mole's tests and generating coverage profile for ${commit_id}" &&  \
    $GO test github.com/${GITHUB_REPOSITORY}/... -v -race -coverprofile=${COV_PROFILE} -covermode=atomic || return 1

  $GO tool cover -html=${COV_PROFILE} -o ${COV_REPORT} || {
    log "error" "error generating coverage report"
    return 1
  }

  log "info" "comparing coverage between ${commit_id} and ${prev_commit_id}"
  cov_diff "$prev_commit_id"
  [ $? -eq 1 ] && return 78

  return 0
}

mole_test
