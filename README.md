# GOOS=libc

```
cd example/helloworld
go run ../../cmd/gooslibc -o helloworld.a .
gcc -o helloworld main.c helloworld.a
./helloworld
```
