#!/bin/bash

IMAGE_NAME=vulcan-stream
DOCKER_REPO=adevinta/vulcan-stream
COMMIT=${TRAVIS_COMMIT:0:7}     # Keep only the 7 leading chars of the commit

docker build \
    --build-arg BUILD_RFC3339=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
    --build-arg COMMIT=$COMMIT \
    -t $IMAGE_NAME .

function add_tag() {
    echo "[***] Adding tag $DOCKER_REPO:$1"
    docker tag $IMAGE_NAME $DOCKER_REPO:$1
    docker push $DOCKER_REPO:$1
}

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

# Add always a tag with the commit id.
add_tag $COMMIT

echo "BRANCH: $TRAVIS_BRANCH"
if [ ! -z $TRAVIS_BRANCH ]; then
    if [[ "${TRAVIS_BRANCH}" == "master" ]] ; then
        BRANCH=latest
    else
        BRANCH=$TRAVIS_BRANCH
    fi
    add_tag $TRAVIS_BRANCH
fi

echo "TAG: $TRAVIS_TAG"
if [ ! -z $TRAVIS_TAG ]; then
    TAG=$TRAVIS_TAG
    if [ "$TAG" =~ "v[0-9]+.*" ]; then
        TAG=${TAG:1}   # Remove the leading "v" from the git tag (ex v1.2 -> 1.3)
    fi
    add_tag $TAG
fi
