# Hitsumabushi (ひつまぶし)

Package hitsumabushi provides APIs to generate JSON for go-build's `-overlay` option.
Hitsumabushi aims to make Go programs work on almost everywhere by overwriting system calls with C function calls.
Now the generated JSON works only for Arm64 so far.

Go version: 1.17.4

## Example

On Arm Linux, run these commands:

```
cd example/helloworld
./build.sh
```
