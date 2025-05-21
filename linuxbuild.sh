#! /bin/bash

docker run --platform linux/amd64 --rm -v "$PWD":/usr/src/visitlog -w /usr/src/visitlog golang:1.24 go build
