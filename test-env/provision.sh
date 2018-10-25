#!/bin/bash

KEYGEN=$(which ssh-keygen)
DOCKER=$(which docker)

APP_PATH=$(dirname $0)
KEY_PATH="${APP_PATH}/key"

[ -z "$KEYGEN" ] && {
  printf "you need ssh-keygen installed to run this tool\n"
  exit 1
}

[ -z "$DOCKER" ] && {
  printf "you need docker installed to run this tool\n"
  exit 1
}

[ ! -f "$KEY_PATH" ] && {
  printf "generating a new ssh key\n"
  $KEYGEN -t rsa -N "" -f $KEY_PATH
}

${DOCKER} cp "${KEY_PATH}.pub" mole_ssh:/home/mole/.ssh/authorized_keys
${DOCKER} exec mole_ssh chown mole:mole /home/mole/.ssh/authorized_keys
${DOCKER} exec mole_ssh chmod 600 /home/mole/.ssh/authorized_keys
