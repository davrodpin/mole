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

install_gomplate() {
  curl -sSL -o ${GOMPLATE} 'https://github.com/hairyhenderson/gomplate/releases/download/v3.3.0/gomplate_linux-amd64-slim' && \
    chmod +x ${GOMPLATE} || return 1

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

cov_diff_data() {
  prev="$1"
  cur="$2"

  [ ! -f "$COV_REPORT" ] && {
    log "error" "coverage diff can't be computed: report file is missing: ${COV_REPORT}"
    return 1
  }

  [ -z "$COV_DIFF_DATA" ] && {
    log "error" "coverage diff output file not defined"
    return 1
  }

  prev_report=`download_report "${prev}"`
  [ $? -ne 0 ] && {
    log "warn" "coverage diff can't be computed: report for ${prev} could not be donwloaded"
    printf "%s\n" "${prev_report}"
    return 2
  }

  cur_stats_data=`cat ${COV_REPORT} | grep "<option value=" | sed -n 's/[[:blank:]]\{0,\}<option value="file[0-9]\{1,\}">\(.\{1,\}\) (\([0-9.]\{1,\}\)%)<\/option>/\1,\2/p'`
  prev_stats_data=`echo "${prev_report}" | grep "<option value=" | sed -n 's/[[:blank:]]\{0,\}<option value="file[0-9]\{1,\}">\(.\{1,\}\) (\([0-9.]\{1,\}\)%)<\/option>/\1,\2/p'`

  [ -z "$cur_stats_data" ] || [ -z "$prev_stats_data" ] && {
    log "error" "could not extract the code coverage numbers from ${COV_REPORT} and/or ${prev}"
    return 1
  }

  now=`date --utc --iso-8601=sec`
  report="$COV_DIFF_DATA_TPL"
  report=`printf "$report\n" | sed "s/{{TITLE}}/Code coverage comparison between ${prev} and ${cur}/"`
  report=`printf "$report\n" | sed "s/{{PREV_COMMIT}}/${prev}/"`
  report=`printf "$report\n" | sed "s/{{CURR_COMMIT}}/${cur}/"`
  report=`printf "$report\n" | sed "s/{{CREATED_AT}}/${now}/"`

  rows=""
  for cur_stats in `printf "%s\n" "$cur_stats_data"`
  do
    file1=`printf "%s\n" "$cur_stats" | awk -F, '{print $1}'`
    cov1=`printf "%s\n" "$cur_stats" | awk -F, '{printf "%.2f", $2}'`
    cov2=0.0
    diff=0.0

    for prev_stats in `printf "%s\n" "$prev_stats_data"`
    do
      file2=`printf "${prev_stats}\n" | awk -F, '{print $1}'`
      cov2=`printf "${prev_stats}\n" | awk -F, '{printf "%.2f", $2}'`

      [ "$file1" = "$file2" ] && {
        diff=`printf "%.2f - %.2f\n" "$cov1" "$cov2" | bc`
        diff=`printf "%.2f" "${diff}"`
        break
      }
    done

    row="$COV_DIFF_DATA_ROW_TPL"
    row=`printf "$row\n" | sed "s:{{FILE}}:${file1}:"`
    row=`printf "$row\n" | sed "s:{{PREV}}:${cov2}:"`
    row=`printf "$row\n" | sed "s:{{CUR}}:${cov1}:"`
    row=`printf "$row\n" | sed "s:{{DIFF}}:${diff}:"`

    if [ -n "$rows" ]
    then
      rows="$rows,$row"
    else
      rows="$row"
    fi
  done

  report=`printf "$report\n" | sed "s#{{ROWS}}#${rows}#"`

  printf "${report}\n" > $COV_DIFF_DATA

  return 0
}

cov_diff_report() {
  prev="$1"
  cur="$2"

  install_gomplate || {
    log "error" "could not download gomplate"
    return 1
  }

  cov_diff_data "$prev" "${cur}"
  ret=$?
  [ $ret -gt 0 ] && return $ret

  [ ! -f "$COV_DIFF_DATA" ] && {
    log "error" "code coverage comparison data could not be found: $COV_DIFF_DATA"
    return 1
  }

  [ ! -f "$COV_DIFF_HTML_REPORT_TPL" ] && {
    log "error" "code coverage report template could not be found: $COV_DIFF_HTML_REPORT_TPL"
    return 1
  }

  [ ! -f "$COV_DIFF_TXT_REPORT_TPL" ] && {
    log "error" "code coverage report template could not be found: $COV_DIFF_TXT_REPORT_TPL"
    return 1
  }


  $GOMPLATE -f ${COV_DIFF_HTML_REPORT_TPL} --context mole=${COV_DIFF_DATA} -o ${COV_DIFF_REPORT} || return 1
  $GOMPLATE -f ${COV_DIFF_TXT_REPORT_TPL} --context mole=${COV_DIFF_DATA} || return 1

  return 0
}

publish() {
  local_path="$1"
	remote_path="$2"

  [ -z "$local_path" ] || [ -z "$remote_path" ] && {
    log "error" "could not publish new report ${local_path} to ${remote_path}"
    return 1
  }

	resp=`curl --silent --show-error -X POST https://content.dropboxapi.com/2/files/upload \
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

  link=`printf "%s" "$resp" | jq '.url' | sed 's/"//g'`

  printf "http://htmlpreview.github.io/?${link}&raw=1"

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

  ## REPORTS

  log "info" "generating report for code coverage comparison between ${prev_commit_id} and ${commit_id}"
  cov_diff_report "${prev_commit_id}" "${commit_id}"

  log "info" "publishing new coverage report for commit ${commit_id}"
  ret=`publish "${COV_REPORT}" "/reports/${commit_id}/mole-coverage.html"`:w
  if [ $? -eq 0 ]
  then
    log "info" "new coverage report available at ${ret}"
  else
    printf "${ret}\n"
  fi

  log "info" "publishing new code coverage comparison report"
  ret=`publish "${COV_DIFF_REPORT}" "/reports/${commit_id}/molde-diff-coverage.html"`
  if [ $? -eq 0 ]
  then
    log "info" "new coverage comparison report available at ${ret}"
  else
    printf "${ret}\n"
  fi


  #TODO publish list of files failing on go-fmt

  #TODO Use $GITHUB_REF to post comment back to PR
  #TODO warning if code coverage decreases

  return 0
}

mole_test
#TODO return 1 if error and 78 if check fails

