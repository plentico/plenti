#!/usr/bin/env bash

if [ -n "$DOCKER_USERNAME" ] && [ -n "$DOCKER_PASSWORD" ]; then
    echo "Login to hub.docker.com"
    echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin
fi