set -e

overlayJSON=$(go run genoverlayjson.go)
env GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
    CGO_CFLAGS='-fno-common -fno-short-enums -ffunction-sections -fdata-sections -fPIC -g -O3' \
    go build -buildmode=c-archive -overlay=$overlayJSON -o=helloworld.a -v
gcc -o helloworld main.c helloworld.a -lpthread 
./helloworld
