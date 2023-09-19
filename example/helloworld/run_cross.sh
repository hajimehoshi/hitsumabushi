# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2023 Hajime Hoshi

set -e

CC="aarch64-linux-gnu-gcc"
go run genoverlayjson.go > /tmp/overlay.json
env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=$CC \
    CGO_CFLAGS='-fno-common -fno-short-enums -ffunction-sections -fdata-sections -fPIC -g -O3' \
    go build -buildmode=c-archive -overlay=/tmp/overlay.json -o=helloworld.a
$CC -o helloworld main.c helloworld.a -lpthread 
env QEMU_LD_PREFIX=/usr/aarch64-linux-gnu qemu-aarch64 ./helloworld
