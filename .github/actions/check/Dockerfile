FROM debian:9.7

LABEL "com.github.actions.name"="action-check"
LABEL "com.github.actions.description"="Code quality check for Mole"
LABEL "com.github.actions.icon"="award"
LABEL "com.github.actions.color"="orange"

LABEL "repository"="http://github.com/davrodpin/mole"
LABEL "homepage"="https://davrodpin.github.io/mole/"
LABEL "maintainer"="David Pinheiro <davrodpin@gmail.com>"

RUN apt-get update && apt install -y curl build-essential jq bc

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
