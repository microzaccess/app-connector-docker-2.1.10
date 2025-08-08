#!/bin/bash

VERSION=2.1.10
DOCKERHUB_USER=microzaccess

# Build and push ztun-watcher
docker build -t $DOCKERHUB_USER/ztun-watcher:$VERSION -f ztun-watcher/Dockerfile .
docker tag $DOCKERHUB_USER/ztun-watcher:$VERSION $DOCKERHUB_USER/ztun-watcher:latest
docker push $DOCKERHUB_USER/ztun-watcher:$VERSION
docker push $DOCKERHUB_USER/ztun-watcher:latest

# Build and push logtrimmer
docker build -t $DOCKERHUB_USER/logtrimmer:$VERSION -f logtrimmer/Dockerfile .
docker tag $DOCKERHUB_USER/logtrimmer:$VERSION $DOCKERHUB_USER/logtrimmer:latest
docker push $DOCKERHUB_USER/logtrimmer:$VERSION
docker push $DOCKERHUB_USER/logtrimmer:latest

# Build and push fluent-bit
docker build -t $DOCKERHUB_USER/fluent-bit:$VERSION -f fluent-bit/Dockerfile .
docker tag $DOCKERHUB_USER/fluent-bit:$VERSION $DOCKERHUB_USER/fluent-bit:latest
docker push $DOCKERHUB_USER/fluent-bit:$VERSION
docker push $DOCKERHUB_USER/fluent-bit:latest
