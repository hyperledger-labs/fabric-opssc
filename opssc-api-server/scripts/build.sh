#! /bin/sh
#
# Copyright 2019-2023 Hitachi, Ltd., Hitachi America, Ltd. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

set -e

if [ "$1" = "" ]; then
  echo "Usage:"
  echo './build.sh $BASE_VERSION $FABRIC_TWO_DIGIT_VERSION'
  exit 1
fi

BASE_VERSION=$1
FABRIC_TWO_DIGIT_VERSION=$2

NODE_VERSION=16
if [ "${FABRIC_TWO_DIGIT_VERSION}" = "2.2" ]; then
  FABRIC_VERSION=2.2.11
else
  FABRIC_VERSION=2.5.1
  FABRIC_TWO_DIGIT_VERSION=2.5
fi

WORK_DIR=$(dirname "$0")
PATH_TO_DOCKER_FILE=${WORK_DIR}/../../Dockerfile-for-api-server
IMAGE_NAME=opssc-api-server
IMAGE_TAG=${BASE_VERSION}-for-fabric-${FABRIC_TWO_DIGIT_VERSION}
LOCAL_IMAGE_PREFIX=fabric-opssc
PUSH_TO_REMOTE=false

docker build -t ${LOCAL_IMAGE_PREFIX}/${IMAGE_NAME}:${IMAGE_TAG} --build-arg FABRIC_VERSION=${FABRIC_VERSION} --build-arg NODE_VERSION=${NODE_VERSION} -f "${PATH_TO_DOCKER_FILE}" "${WORK_DIR}/../.."
if [ "${FABRIC_TWO_DIGIT_VERSION}" = "2.5" ]; then
  docker tag ${LOCAL_IMAGE_PREFIX}/${IMAGE_NAME}:${IMAGE_TAG} ${LOCAL_IMAGE_PREFIX}/${IMAGE_NAME}:latest
fi

if [ "$PUSH_TO_REMOTE" = "true" ]; then
  docker tag ${LOCAL_IMAGE_PREFIX}/${IMAGE_NAME}:${IMAGE_TAG} "${REGISTRY_HOST}"/"${REGISTRY_NS}"/${IMAGE_NAME}:${IMAGE_TAG}
  docker push "${REGISTRY_HOST}"/"${REGISTRY_NS}"/${IMAGE_NAME}:${IMAGE_TAG}
fi
