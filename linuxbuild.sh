#! /bin/bash

docker run --rm -v "$PWD":/usr/src/visitlog -w /usr/src/visitlog golang:1.8 go build

