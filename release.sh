#!/bin/bash

rm -rf .build .tarballs

VERSION=$(cat VERSION)

echo "Version : $VERSION"

make build
promu crossbuild
promu crossbuild tarballs
promu checksum .tarballs
promu release .tarballs

rm solr_exporter
ln -s .build/linux-amd64/solr_exporter solr_exporter

make docker DOCKER_IMAGE_NAME=stanchan/prometheus-solr-exporter DOCKER_IMAGE_TAG=v$VERSION
docker login
docker tag "stanchan/prometheus-solr-exporter:v$VERSION" "stanchan/prometheus-solr-exporter:latest"
docker push stanchan/prometheus-solr-exporter