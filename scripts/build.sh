#!/usr/bin/bash

ORG_NAME="tampler"
IMG_NAME="nws-api-gateway"

docker build --build-arg SSH_PRIV_KEY="$(cat ~/.ssh/id_rsa)" -f Dockerfile.dev -t ${ORG_NAME}/${IMG_NAME} .
