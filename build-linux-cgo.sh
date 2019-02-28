#!/usr/bin/env bash
export GO111MODULE=on

VERSION=$(git describe --tags)
BUILD_TIME=`date +%FT%T%z`

xgo -ldflags="-w -X main.BuildTime=$BUILD_TIME -X main.Version=$VERSION" --targets="linux/386,linux/amd64,linux/arm-6,linux/arm-7,linux/arm64" .

VERSION=${VERSION#v}

mv standardfile-linux-386 standardfile
tar czf standardfile_${VERSION}_linux_32-bit.tar.gz standardfile
rm -f standardfile

mv standardfile-linux-amd64 standardfile
tar czf standardfile_${VERSION}_linux_64-bit.tar.gz standardfile
rm -f standardfile

mv standardfile-linux-arm-6 standardfile
tar czf standardfile_${VERSION}_linux_arm6.tar.gz standardfile
rm -f standardfile

mv standardfile-linux-arm-7 standardfile
tar czf standardfile_${VERSION}_linux_arm7.tar.gz standardfile
rm -f standardfile

mv standardfile-linux-arm64 standardfile
tar czf standardfile_${VERSION}_linux_arm8.tar.gz standardfile
rm -f standardfile

mv standardfile_${VERSION}* dist/