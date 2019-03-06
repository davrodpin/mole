#!/bin/sh -l

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

  printf "%s" "$resp" | jq '.url' | sed 's/"//g'

  return 0
}

mole_publish() {
  [ ! -f "$COV_REPORT" ] && {
    log "error" "coverage report could not be found (${COV_REPORT})"
    return 1
  }

  commit_id=`jq '.after' ${GITHUB_EVENT_PATH} | sed 's/"//g' | cut -c-7`

  log "info" "publishing new coverage report ${remote_path} for commit ${commit_id}"
  link=`publish "${COV_REPORT}" "/reports/${commit_id}/mole-coverage.html"`
  retcode=$?

  [ $retcode -ne 0 ] && {
    log "error" "new coverage report could not be published"
    printf "${link}"
  }

  report_url="http://htmlpreview.github.io/?${link}&raw=1"
  log "info" "coverage report available on ${report_url}"

  return $retcode
}

mole_publish

