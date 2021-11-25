# GOOS=libc

On Arm Linux, run these commands:

```
cd example/helloworld
go run ../../cmd/gooslibc -o helloworld.a .
gcc -o helloworld main.c helloworld.a
./helloworld
```
