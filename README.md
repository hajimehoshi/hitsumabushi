# Hitsumabushi (ひつまぶし)

Package hitsumabushi provides APIs to generate JSON for go-build's `-overlay` option.
Hitsumabushi aims to make Go programs work on almost everywhere by overwriting system calls with C function calls.
Now the generated JSON works only for Linux/Amd64, Linux/Arm64, and Windows/Amd64 so far.
For GOOS=windows, Hitsumabushi replaces some functions that don't work on some special Windows-like systems.

Go version: 1.19-1.24

## Example

On Arm Linux, run these commands:

```
cd example/helloworld
./run.sh
```

## Tips

With VC++, you might have to call `_rt0_amd64_windows_lib()` at the beginning of the entry point explicitly.
See also https://github.com/golang/go/issues/42190.
