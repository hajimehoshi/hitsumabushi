# SPDX-License-Identifier: Apache-2.0

set -e

go run ./genoverlayjson.go test $@ > /tmp/overlay.json
env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
    CGO_CFLAGS='-fno-common -fno-short-enums -ffunction-sections -fdata-sections -fPIC -g -O3' \
    go test -c -vet=off -overlay=/tmp/overlay.json -o=test ${@: -1}
./test
