FROM debian:9.7

LABEL "com.github.actions.name"="action-report"
LABEL "com.github.actions.description"="Create quality reports for Mole"
LABEL "com.github.actions.icon"="archive"
LABEL "com.github.actions.color"="blue"

LABEL "repository"="http://github.com/davrodpin/mole"
LABEL "homepage"="https://davrodpin.github.io/mole/"
LABEL "maintainer"="David Pinheiro <davrodpin@gmail.com>"

RUN apt-get update && apt install -y curl build-essential jq bc

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
