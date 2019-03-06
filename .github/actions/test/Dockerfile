FROM debian:9.7

LABEL "com.github.actions.name"="action-test"
LABEL "com.github.actions.description"="Execute Mole's tests"
LABEL "com.github.actions.icon"="box"
LABEL "com.github.actions.color"="purple"

LABEL "repository"="http://github.com/davrodpin/mole"
LABEL "homepage"="https://davrodpin.github.io/mole/"
LABEL "maintainer"="David Pinheiro <davrodpin@gmail.com>"

RUN apt-get update && apt install -y curl build-essential jq bc

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
