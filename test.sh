# SPDX-License-Identifier: Apache-2.0

set -e

overlayJSON=$(go run ./example/helloworld/genoverlayjson.go)
env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
    CGO_CFLAGS='-fno-common -fno-short-enums -ffunction-sections -fdata-sections -fPIC -g -O3' \
    go test -c -vet=off -overlay=$overlayJSON -o=test runtime
./test -test.v -test.short
