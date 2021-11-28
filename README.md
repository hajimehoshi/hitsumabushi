# GOOS=libc

Go version: 1.17.3

On Arm Linux, run these commands:

```
cd example/helloworld
go run ../../cmd/gooslibc -o helloworld.a .
gcc -o helloworld main.c helloworld.a -lpthread
./helloworld
```
