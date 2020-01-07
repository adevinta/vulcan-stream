#!/bin/bash

IMAGE_NAME=vulcan-stream
DOCKER_REPO=adevinta/vulcan-stream

docker build \
    --build-arg BUILD_RFC3339=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
    --build-arg COMMIT=$1 \
    -t $IMAGE_NAME .

function add_tag() {
    echo "[***] Adding tag $DOCKER_REPO:$1"
    docker tag $IMAGE_NAME $DOCKER_REPO:$1
    docker push $DOCKER_REPO:$1
}

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

# Add always a tag with the commit id.
add_tag $1

echo "BRANCH: $2"
if [[ -z "$2" ]];
then
    if [[ "$2" == "master" ]] ; then
        BRANCH=latest
    fi
    add_tag $2
fi

echo "TAG: $3"
if [[ -z "$3" ]];
then
  TAG="${3:1}"  # remove leading 'v'
  add_tag $TAG
fi
