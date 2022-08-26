# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2022 Hajime Hoshi

set -e

go run genoverlayjson.go > /tmp/overlay.json
env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
    CGO_CFLAGS='-fno-common -fno-short-enums -ffunction-sections -fdata-sections -fPIC -g -O3' \
    go build -buildmode=c-archive -overlay=/tmp/overlay.json -o=helloworld.a
gcc -o helloworld main.c helloworld.a -lpthread 
./helloworld
