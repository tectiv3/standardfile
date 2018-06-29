#!/usr/bin/env bash

version=$(./standardfile -v)

xgo --targets="linux/386,linux/amd64,linux/arm-6,linux/arm-7,linux/arm64" .

mv standardfile-linux-386 standardfile
tar czf standardfile_${version}_linux_32-bit.tgz standardfile
rm -f standardfile

mv standardfile-linux-amd64 standardfile
tar czf standardfile_${version}_linux_64-bit.tgz standardfile
rm -f standardfile

mv standardfile-linux-arm-6 standardfile
tar czf standardfile_${version}_linux_arm6.tgz standardfile
rm -f standardfile

mv standardfile-linux-arm-7 standardfile
tar czf standardfile_${version}_linux_arm7.tgz standardfile
rm -f standardfile

mv standardfile-linux-arm64 standardfile
tar czf standardfile_${version}_linux_arm8.tgz standardfile
rm -f standardfile

mv standardfile_${version}* dist/