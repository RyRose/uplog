#!/usr/bin/env bash

set -eux

git tag "${TAG}"
docker build . -t "uplog/uplog:${TAG}"
docker push "uplog/uplog:${TAG}"
