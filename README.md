# Hitsumabushi (ひつまぶし)

Hitsumabushi provides a JSON file for go-build's `-overlay` option in order to run Go programs (almost) everywhere.

Go version: 1.17.4

On Arm Linux, run these commands:

```
cd example/helloworld
go run ../../cmd/gooslibc -o helloworld.a .
gcc -o helloworld main.c helloworld.a -lpthread
./helloworld
```
