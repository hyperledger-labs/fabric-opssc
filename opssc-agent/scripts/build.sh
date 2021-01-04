#! /bin/sh
#
# Copyright 2019, 2020 Hitachi America, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -e
WORK_DIR=$(dirname "$0")
PATH_TO_DOCKER_FILE=${WORK_DIR}/../../Dockerfile-for-agent
IMAGE_NAME=opssc-agent
IMAGE_TAG=latest
LOCAL_IMAGE_PREFIX=fabric-opssc
PUSH_TO_REMOTE=false

docker build -t ${LOCAL_IMAGE_PREFIX}/${IMAGE_NAME}:${IMAGE_TAG} -f ${PATH_TO_DOCKER_FILE} ${WORK_DIR}/../..

if [ "$PUSH_TO_REMOTE" = "true" ]; then
  docker tag ${LOCAL_IMAGE_PREFIX}/${IMAGE_NAME}:${IMAGE_TAG} "${REGISTRY_HOST}"/"${REGISTRY_NS}"/${IMAGE_NAME}:${IMAGE_TAG}
  docker push "${REGISTRY_HOST}"/"${REGISTRY_NS}"/${IMAGE_NAME}:${IMAGE_TAG}
fi
