#!/bin/bash
local_arch=$(uname -m)
arch=${1-$local_arch}

if [ "$arch" = "aarch64" ]; then
    echo "aarch64"
    go_arch="arm64"
elif [ "$arch" = "x86_64" ]; then
    echo "x86_64"
    go_arch="amd64"
else
    echo "不支持的架构"
    exit 1
fi
echo build $go_arch...
out=kubeconfigmerge-linux-$go_arch
GOOS=linux GOARCH=$go_arch go build -o $out kubeconfigmerge.go 
echo build $go_arch done: $out