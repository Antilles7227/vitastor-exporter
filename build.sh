#!/usr/bin/env bash
archs=(386 amd64 arm arm64)
oss=(linux darwin)
version=$1
for arch in ${archs[@]} 
do
    for os in ${oss[@]}
    do
        echo "Building exporter, OS: $os, Arch: $arch"
        env GOOS=${os} GOARCH=${arch} go build -o bin/vitastor-exporter-${version}-${os}-${arch}
    done
done
