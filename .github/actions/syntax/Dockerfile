FROM debian:9.7

LABEL "com.github.actions.name"="action-syntax"
LABEL "com.github.actions.description"="Validates Go source code syntax"
LABEL "com.github.actions.icon"="file-text"
LABEL "com.github.actions.color"="orange"

LABEL "repository"="http://github.com/davrodpin/mole"
LABEL "homepage"="https://davrodpin.github.io/mole/"
LABEL "maintainer"="David Pinheiro <davrodpin@gmail.com>"

RUN apt-get update && apt install -y curl build-essential

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
