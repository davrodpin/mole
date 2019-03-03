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

commit() {
  path="$1"
  message="$2"
  content="$3"

  [ -z "$path" ] || [ -z "$message" ] || [ -z "$content" ] && {
    log "info" "could not publish new report"
    return 1
  }

	payload=`printf "$JSON" | sed "s:{{MESSAGE}}:${message}:"`
	payload=`printf "$payload" | sed "s:{{CONTENT}}:${content}:"`

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

mole_test() {
  go_install && \
    mole_create_wksp && \
    log "info" "running mole's tests..." &&  \
    $GO test github.com/${GITHUB_REPOSITORY}/... -v -race -coverprofile=${COV_PROFILE} -covermode=atomic || return 1

  return 0
}

get_file_content() {
  path="$1"
  ref="$2"
  debug="$3"

  [ -z "$path" ] && return 2

  [ -z "$ref" ] && ref="master"

  uri="/repos/davrodpin/mole/contents${path}"
  url="https://api.github.com${uri}?ref=${ref}"

  resp=`curl -H "Authorization: token ${GITHUB_TOKEN}" "${url}" | jq '.content' | sed 's/"//g'`

  [ -z "$response" ] && return 3

  content=`echo "${resp}" | base64 -d -`

  [ -n "${debug}" ] && {
    printf "url: ${url}\n"
    printf ">>> response:\n"
    printf "%s\n\n" "${resp}"
    printf ">>> content:\n"
    printf "%s\n\n" "${content}"
  }

  [ -z "$content" ]  && return 4

  echo "${content}"

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

  #printf "\n---------- DEBUG ----------\n"
  #get_file_content "/docs/cov/${that}/mole-coverage.html" "master" "debug:on"
  #printf "\n---------- END OF DEBUG ----------\n"
  that_report=`get_file_content "/docs/cov/${that}/mole-coverage.html" "master"`
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

  jq '.' ${GITHUB_EVENT_PATH}

  prev_commit_id=`jq '.before' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`
  commit_id=`jq '.after' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`
  report_name="mole-coverage.html"
  report_path=${MOLE_SRC_PATH}/docs/cov/${commit_id} #${GITHUB_WORKSPACE}/${report_name}

  mkdir -p ${report_path} && $GO tool cover -html=${COV_PROFILE} -o ${report_path}/${report_name} || {
    log "error" "error generating coverage report"
    return 1
  }

  path="docs/cov/${commit_id}/${report_name}"
	message="Add new coverage report"
  content=`base64 -w0 ${report_path}/${report_name}`

  commit "$path" "$message" "$content"
  retcode=$?

  [ $retcode -eq 0 ] && {
    report_url="http://htmlpreview.github.io/?https://github.com/${GITHUB_REPOSITORY}/blob/master/${path}"
    log "info" "coverage report available on ${report_url}"
  }

  log "info" "comparing coverage between ${commit_id} and ${prev_commit_id}"
  compare "$commit_id" "$prev_commit_id"

  return $retcode
}

mole_report
