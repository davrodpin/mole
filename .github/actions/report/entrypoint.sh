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

publish() {
  local_path="$1"
	remote_path="$2"

  [ -z "$local_path" ] || [ -z "$remote_path" ] && {
    log "error" "could not publish new report ${local_path} to ${remote_path}"
    return 1
  }

	reponse=`curl --silent --show-error -X POST https://content.dropboxapi.com/2/files/upload \
    --header "Authorization: Bearer ${DROPBOX_TOKEN}" \
    --header "Dropbox-API-Arg: {\"path\":\"${remote_path}\", \"mode\":\"overwrite\", \"mute\":true}" \
    --header "Content-Type: application/octet-stream" \
    --data-binary @${local_path}`

  error=`printf "%s\n" "$resp" | jq 'select(.error != null) | .error'`
  [ -n "$error" ] && {
    log "error" "could not publish report ${local_path}"
    printf "%s\n" "$resp" | jq '.'
    return 1
  }

  resp=`curl --silent --show-error -X POST https://api.dropboxapi.com/2/sharing/create_shared_link_with_settings \
    --header "Authorization: Bearer ${DROPBOX_TOKEN}" \
    --header "Content-Type: application/json" \
    --data "{\"path\": \"${remote_path}\",\"settings\": {\"requested_visibility\": \"public\"}}"`

  printf "%s\n" "$resp" | grep -q 'Error in call' && {
    log "error" "report was published but could not create public link for ${remote_path}: ${resp}"
    return 1
  }

  printf "%s" "$resp" | jq '.url' | sed 's/"//g'

  return 0
}

mole_test() {
  go_install && \
    mole_create_wksp && \
    log "info" "running mole's tests..." &&  \
    $GO test github.com/${GITHUB_REPOSITORY}/... -v -race -coverprofile=${COV_PROFILE} -covermode=atomic || return 1

  return 0
}

download_report() {
  commit="$1"

  [ -z "$commit" ] && return 2

  resp=`curl --silent --show-error -X POST https://content.dropboxapi.com/2/files/download \
    --header "Authorization: Bearer ${DROPBOX_TOKEN}" \
    --header "Dropbox-API-Arg: {\"path\": \"/reports/${commit}/mole-coverage.html\"}"`

  error=`printf "%s\n" "$resp" | jq 'select(.error != null) | .error' 2> /dev/null`
  [ -n "$error" ] && {
    log "error" "could not fetch coverage report for ${commit}"
    printf "%s\n" "${resp}"
    return 1
  }

  printf "%s\n" "${resp}"

  return 0
}

compare() {
  this="$1"
  that="$2"

  this_report=`ls ${MOLE_SRC_PATH}/docs/cov/${this}/*.html | tail -n1`

  [ -z "$this" ] && {
    log "error" "report file is missing: ${this_report}"
    return 1
  }

  that_report=`download_report "${that}"`
  case $? in
    2)
      log "error" "can't get file content with empty path"
      return 1
      ;;
    3)
      log "error" "error fetching file content for ${that}"
      return 1
      ;;
    4)
      log "error" "error decoding file content for ${that}"
      return 1
      ;;
  esac

  this_stats=`cat ${this_report} | grep "<option value=" | sed -n 's/[[:blank:]]\{0,\}<option value="file[0-9]\{1,\}">\(.\{1,\}\) (\([0-9.]\{1,\}\)%)<\/option>/\1,\2/p'`
  that_stats=`echo "${that_report}" | grep "<option value=" | sed -n 's/[[:blank:]]\{0,\}<option value="file[0-9]\{1,\}">\(.\{1,\}\) (\([0-9.]\{1,\}\)%)<\/option>/\1,\2/p'`

  [ -z "$this_stats" ] || [ -z "$that_stats" ] && {
    log "error" "could not extract the code coverage numbers from ${this}:${this_report} and ${that}"
    return 1
  }

  for stats1 in `printf "%s\n" "$this_stats"`
  do
    mod1=`printf "%s\n" "$stats1" | awk -F, '{print $1}'`
    cov1=`printf "%s\n" "$stats1" | awk -F, '{print $2}'`
    diff=0

    for stats2 in `printf "%s\n" "$that_stats"`
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

mole_report() {
  go_install || return 1

  [ ! -f "$COV_PROFILE" ] && {
    mole_test 2>&1 > /dev/null || return 1
  }

  go_install || return 1

  prev_commit_id=`jq '.before' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`
  commit_id=`jq '.after' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`
  report_name="mole-coverage.html"
  report_path=${MOLE_SRC_PATH}/docs/cov/${commit_id}

  mkdir -p ${report_path} && $GO tool cover -html=${COV_PROFILE} -o ${report_path}/${report_name} || {
    log "error" "error generating coverage report"
    return 1
  }

  log "info" "publishing new coverage report ${remote_path}"
  link=`publish "${report_path}/${report_name}" "/reports/${commit_id}/${report_name}"`
  retcode=$?

  [ $retcode -ne 0 ] && {
    log "error" "new coverage report could not be published"
    printf "${link}"
  }

  report_url="http://htmlpreview.github.io/?${link}&raw=1"
  log "info" "coverage report available on ${report_url}"

  log "info" "comparing coverage between ${commit_id} and ${prev_commit_id}"
  compare "$commit_id" "$prev_commit_id"

  return $retcode
}

mole_report
