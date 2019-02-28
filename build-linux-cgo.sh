#!/usr/bin/env bash
export GO111MODULE=on

VERSION=$(git describe --tags)
BUILD_TIME=`date +%FT%T%z`

xgo -ldflags="-w -X main.BuildTime=$BUILD_TIME -X main.Version=$VERSION" --targets="linux/386,linux/amd64,linux/arm-6,linux/arm-7,linux/arm64" .

mkdir dist/linux_386
mv standardfile-linux-386 dist/linux_386/standardfile

mkdir dist/linux_amd64
mv standardfile-linux-amd64 dist/linux_amd64/standardfile

mkdir dist/linux_arm6
mv standardfile-linux-arm-6 dist/linux_arm6/standardfile

mkdir dist/linux_arm7
mv standardfile-linux-arm-7 dist/linux_arm7/standardfile

mkdir dist/linux_arm8
mv standardfile-linux-arm64 dist/linux_arm8/standardfile
