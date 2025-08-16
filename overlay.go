// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 Hajime Hoshi

package hitsumabushi

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf16"
)

type Option func(*config)

type replaceString struct {
	from string
	to   string
}

type config struct {
	testPkgs    []string
	overlayDir  string
	os          string
	args        []string
	pageSize    int
	replaceDLLs []replaceString
	overlay     []replaceString
}

// TestPkg represents a package for testing.
// When generating a JSON, importing `runtime/cgo` is inserted in the testing package.
func TestPkg(pkg string) Option {
	return func(cfg *config) {
		cfg.testPkgs = append(cfg.testPkgs, pkg)
	}
}

// OverlayDir sets the temporary working directory where overlay files are stored.
func OverlayDir(dir string) Option {
	return func(cfg *config) {
		cfg.overlayDir = dir
	}
}

// Args is arguments when executing.
// The first argument must be a program name.
func Args(args ...string) Option {
	return func(cfg *config) {
		cfg.args = append(cfg.args, args...)
	}
}

// GOOS specifies GOOS to generate the JSON.
// The default value is runtime.GOOS.
func GOOS(os string) Option {
	return func(cfg *config) {
		cfg.os = os
	}
}

// PageSize specifies the page size.
// The deafult value is 4096.
//
// This works only for Linux.
func PageSize(pageSize int) Option {
	return func(cfg *config) {
		cfg.pageSize = pageSize
	}
}

// ReplaceDLL replaces a DLL file name loaded at LoadLibraryW and LoadLibraryExW.
//
// This works only for Windows.
func ReplaceDLL(from, to string) Option {
	return func(cfg *config) {
		cfg.replaceDLLs = append(cfg.replaceDLLs, replaceString{
			from: from,
			to:   to,
		})
	}
}

// Overlay adds or replaces an entry for the -overlay option.
func Overlay(from, to string) Option {
	return func(cfg *config) {
		cfg.overlay = append(cfg.overlay, replaceString{
			from: from,
			to:   to,
		})
	}
}

//go:embed 1.*_*
var patchFiles embed.FS

// reGoVersion represents a regular expression for Go version.
// With gotip, the version might start with "devel ", so '^' is not used here.
var reGoVersion = regexp.MustCompile(`go(\d+\.\d+)`)

func goVersion() string {
	m := reGoVersion.FindStringSubmatch(runtime.Version())
	return m[1]
}

func goMajorMinorVersion() (int, int) {
	tokens := strings.Split(goVersion(), ".")
	major, _ := strconv.Atoi(tokens[0])
	minor, _ := strconv.Atoi(tokens[1])
	return major, minor
}

// GenOverlayJSON generates a JSON file for go-build's `-overlay` option.
// GenOverlayJSON returns a JSON file content, or an error if generating it fails.
//
// Now the generated JSON works only for Arm64 so far.
func GenOverlayJSON(options ...Option) ([]byte, error) {
	type overlay struct {
		Replace map[string]string
	}

	cfg := config{
		os:       runtime.GOOS,
		pageSize: 4096,
	}
	for _, op := range options {
		op(&cfg)
	}

	root := goVersion() + "_" + cfg.os
	dir, err := patchFiles.Open(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("hitsumabushi: Hitsumabushi does not support the Go version %s and GOOS=%s", runtime.Version(), cfg.os)
		}
		return nil, err
	}
	dir.Close()

	subFiles, err := fs.Sub(patchFiles, root)
	if err != nil {
		return nil, err
	}

	tmpDir := cfg.overlayDir
	if tmpDir == "" {
		var err error
		tmpDir, err = os.MkdirTemp("", "")
		if err != nil {
			return nil, err
		}
	}

	replaces := map[string]string{}
	if err := fs.WalkDir(subFiles, ".", func(entryPath string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		const (
			modTypeReplace = ".replace"
			modTypeAppend  = ".append"
			modTypePatch   = ".patch"
		)
		modType := modTypeReplace

		origFilename := path.Base(entryPath)
		for _, m := range []string{modTypeAppend, modTypePatch} {
			if strings.HasSuffix(origFilename, m) {
				origFilename = origFilename[:len(origFilename)-len(m)]
				modType = m
				break
			}
		}

		ext := path.Ext(origFilename)
		if ext != ".go" && ext != ".c" && ext != ".h" && ext != ".s" {
			return nil
		}

		pkg := path.Dir(entryPath)
		origDir, err := goPkgDir(pkg, cfg.os)
		if err != nil {
			return err
		}

		src, err := subFiles.Open(entryPath)
		if err != nil {
			return err
		}
		defer src.Close()

		if err := os.MkdirAll(filepath.Join(tmpDir, filepath.FromSlash(pkg)), 0755); err != nil {
			return err
		}
		dst, err := os.Create(filepath.Join(tmpDir, filepath.FromSlash(pkg), origFilename))
		if err != nil {
			return err
		}
		defer dst.Close()

		origPath := filepath.Join(origDir, origFilename)
		defer func() {
			replaces[origPath] = dst.Name()
		}()

		switch modType {
		case modTypeReplace:
			if _, err := io.Copy(dst, src); err != nil {
				return err
			}

		case modTypeAppend:
			orig, err := os.Open(origPath)
			if err != nil {
				return err
			}
			defer orig.Close()

			if _, err := io.Copy(dst, io.MultiReader(orig, src)); err != nil {
				return err
			}

		case modTypePatch:
			orig, err := os.Open(origPath)
			if err != nil {
				return err
			}
			defer orig.Close()

			p, err := parsePatch(entryPath, src)
			if err != nil {
				return err
			}
			patched, err := p.apply(orig)
			if err != nil {
				return err
			}
			if _, err := io.Copy(dst, patched); err != nil {
				return err
			}

		default:
			return fmt.Errorf("hitsumabushi: unexpected modType: %s", modType)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	switch cfg.os {
	case "linux":
		// Replace the arguments.
		{
			var strs []string
			for _, arg := range cfg.args {
				strs = append(strs, fmt.Sprintf(`%q`, arg))
			}
			argvDef := "var __argv = [...]string{" + strings.Join(strs, ", ") + "}"

			old := `func goargs() {
	if GOOS == "windows" {
		return
	}
	argslice = make([]string, argc)
	for i := int32(0); i < argc; i++ {
		argslice[i] = gostringnocopy(argv_index(argv, i))
	}
}`
			new := fmt.Sprintf(`%s

func goargs() {
	if GOOS == "windows" {
		return
	}
	argslice = make([]string, %[2]d)
	if len(argslice) == 0 {
		// os.Executable is not available here. Give a dummy name.
		argslice = []string{"hitsumabushi_app"}
	} else {
		for i := int32(0); i < %[2]d; i++ {
			argslice[i] = __argv[i]
		}
	}
}`, argvDef, len(cfg.args))
			if err := replace(tmpDir, replaces, "runtime", "runtime1.go", old, new, cfg.os); err != nil {
				return nil, err
			}
		}

		{
			old := `var physPageSize uintptr`
			new := fmt.Sprintf(`var physPageSize uintptr = %d`, cfg.pageSize)
			if err := replace(tmpDir, replaces, "runtime", "malloc.go", old, new, cfg.os); err != nil {
				return nil, err
			}
		}

	case "windows":
		// Replace the arguments.
		{
			var strs []string
			for _, arg := range cfg.args {
				strs = append(strs, fmt.Sprintf(`%q`, arg))
			}
			argvDef := "var __argv = []string{" + strings.Join(strs, ", ") + "}"

			// It is hard to emulate GetCommandLine exactly.
			// See http://daviddeley.com/autohotkey/parameters/parameters.htm#WINARGV
			// Initialize os.Args directly instead.
			old := `func init() {
	cmd := windows.UTF16PtrToString(syscall.GetCommandLine())
	if len(cmd) == 0 {
		arg0, _ := Executable()
		Args = []string{arg0}
	} else {
		Args = commandLineToArgv(cmd)
	}
}`
			new := fmt.Sprintf(`%s

func init() {
	if len(__argv) == 0 {
		arg0, _ := Executable()
		__argv = []string{arg0}
	}
	Args = __argv
}`, argvDef)
			if err := replace(tmpDir, replaces, "os", "exec_windows.go", old, new, cfg.os); err != nil {
				return nil, err
			}

			major, minor := goMajorMinorVersion()
			if major != 1 {
				return nil, fmt.Errorf("hitsumabushi: unexpected major version: %d", major)
			}
			if minor >= 23 {
				if err := replace(tmpDir, replaces, "os", "exec_windows.go", `import (
	"errors"
	"internal/syscall/windows"
	"runtime"
	"syscall"
	"time"
)`, `import (
	"errors"
	"runtime"
	"syscall"
	"time"
)`, cfg.os); err != nil {
					return nil, err
				}
			} else {
				if err := replace(tmpDir, replaces, "os", "exec_windows.go", `import (
	"errors"
	"internal/syscall/windows"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
)`, `import (
	"errors"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
)`, cfg.os); err != nil {
					return nil, err
				}
			}
		}

		// Replace loaded DLLs
		if len(cfg.replaceDLLs) > 0 {
			var old string
			_, minor := goMajorMinorVersion()
			if minor <= 22 {
				old = "func syscall_SyscallN(trap uintptr, args ...uintptr) (r1, r2, err uintptr) {"
			} else {
				old = "func syscall_SyscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {"
			}
			new := `func _toLower(x uint16) uint16 {
	if 'A' <= x && x <= 'Z' {
		return x - 'A' + 'a'
	}
	return x
}

func _areUTF16StringsSame(a *uint16, b *uint16) bool {
	for _toLower(*a) == _toLower(*b) {
		a = (*uint16)(unsafe.Add(unsafe.Pointer(a), 2))
		b = (*uint16)(unsafe.Add(unsafe.Pointer(b), 2))
		if *a == 0 || *b == 0 {
			return *a == 0 && *b == 0
		}
	}
	return false
}

var _replacingDLLFroms = [][]uint16{
	{{.Froms}}
}

var _replacingDLLTos = [][]uint16{
	{{.Tos}}
}
`
			if minor <= 22 {
				new += `
func syscall_SyscallN(trap uintptr, args ...uintptr) (r1, r2, err uintptr) {
	if trap == getLoadLibrary() || trap == getLoadLibraryEx() {
		for i, from := range _replacingDLLFroms {
			if _areUTF16StringsSame((*uint16)(unsafe.Pointer(args[0])), &from[0]) {
				args[0] = uintptr(unsafe.Pointer(&_replacingDLLTos[i][0]))
				break
			}
		}
	}`
			} else {
				new += `
func syscall_SyscallN(fn uintptr, args ...uintptr) (r1, r2, err uintptr) {
	if fn == uintptr(unsafe.Pointer(_LoadLibraryW)) || fn == uintptr(unsafe.Pointer(_LoadLibraryExW)) {
		for i, from := range _replacingDLLFroms {
			if _areUTF16StringsSame((*uint16)(unsafe.Pointer(args[0])), &from[0]) {
				args[0] = uintptr(unsafe.Pointer(&_replacingDLLTos[i][0]))
				break
			}
		}
	}`
			}
			var froms []string
			var tos []string
			for _, replace := range cfg.replaceDLLs {
				from, err := utf16FromString(replace.from)
				if err != nil {
					return nil, err
				}
				froms = append(froms, fmt.Sprintf("%#v,", from))

				to, err := utf16FromString(replace.to)
				if err != nil {
					return nil, err
				}
				tos = append(tos, fmt.Sprintf("%#v,", to))
			}
			new = strings.ReplaceAll(new, "{{.Froms}}", strings.Join(froms, "\n\t"))
			new = strings.ReplaceAll(new, "{{.Tos}}", strings.Join(tos, "\n\t"))
			if err := replace(tmpDir, replaces, "runtime", "syscall_windows.go", old, new, cfg.os); err != nil {
				return nil, err
			}
		}
	}

	// Add importing "runtime/cgo" for testing packages.
	for _, pkg := range cfg.testPkgs {
		origPath, err := goExternalTestFile(pkg)
		if err != nil {
			return nil, err
		}

		pkgName, err := goPkgName(pkg)
		if err != nil {
			return nil, err
		}

		// Read the source before opening the destination.
		// The destination might be the same as the source.
		srcPath := origPath
		if p, ok := replaces[origPath]; ok {
			srcPath = p
		}
		srcContent, err := os.ReadFile(srcPath)
		if err != nil {
			return nil, err
		}

		if err := os.MkdirAll(filepath.Join(tmpDir, filepath.FromSlash(pkg)), 0755); err != nil {
			return nil, err
		}
		dst, err := os.Create(filepath.Join(tmpDir, filepath.FromSlash(pkg), filepath.Base(origPath)))
		if err != nil {
			return nil, err
		}
		defer dst.Close()

		// This assumes that there is an external test package.
		old := "package " + pkgName + "_test"
		new := old + "\n\n" + `import _ "runtime/cgo"`
		replaced := strings.Replace(string(srcContent), old, new, 1)

		if _, err := io.Copy(dst, bytes.NewReader([]byte(replaced))); err != nil {
			return nil, err
		}

		replaces[origPath] = dst.Name()
	}

	for _, r := range cfg.overlay {
		replaces[r.from] = r.to
	}

	return json.Marshal(&overlay{Replace: replaces})
}

func goPkgDir(pkg string, goos string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", pkg)
	cmd.Env = append(os.Environ(), "GOOS="+goos)
	cmd.Stderr = &buf
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("hitsumabushi: %v\n%s", err, buf.String())
	}
	return strings.TrimSpace(string(out)), nil
}

func goExternalTestFile(pkg string) (string, error) {
	idx := 0
	for {
		var buf bytes.Buffer
		cmd := exec.Command("go", "list", "-f", "{{.Dir}}"+string(filepath.Separator)+fmt.Sprintf("{{index .XTestGoFiles %d}}", idx), pkg)
		cmd.Stderr = &buf
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("hitsumabushi: %v\n%s\nperhaps this package doesn't have an external test", err, buf.String())
		}

		f := strings.TrimSpace(string(out))

		// runtime/callers_test.go is very special and the line number matters.
		if pkg == "runtime" && filepath.Base(f) == "callers_test.go" {
			idx++
			continue
		}

		return f, nil
	}
}

func goPkgName(pkg string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("go", "list", "-f", "{{.Name}}", pkg)
	cmd.Stderr = &buf
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("hitsumabushi: %v\n%s", err, buf.String())
	}
	return strings.TrimSpace(string(out)), nil
}

func replace(tmpDir string, replaces map[string]string, pkg string, filename string, old, new string, goos string) error {
	origDir, err := goPkgDir(pkg, goos)
	if err != nil {
		return err
	}
	origPath := filepath.Join(origDir, filename)

	// Read the source before opening the destination.
	// The destination might be the same as the source.
	srcPath := origPath
	if p, ok := replaces[origPath]; ok {
		srcPath = p
	}
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(tmpDir, filepath.FromSlash(pkg)), 0755); err != nil {
		return err
	}
	dst, err := os.Create(filepath.Join(tmpDir, filepath.FromSlash(pkg), filepath.Base(origPath)))
	if err != nil {
		return err
	}
	defer dst.Close()

	if !strings.Contains(string(srcContent), old) {
		return fmt.Errorf("hitsumabushi: replacing %s/%s failed: %s", pkg, filename, old)
	}

	replaced := strings.Replace(string(srcContent), old, new, 1)
	if string(srcContent) == replaced {
		return fmt.Errorf("hitsumabushi: replacing %s/%s failed: replacing result is the same", pkg, filename)
	}
	if _, err := io.Copy(dst, bytes.NewReader([]byte(replaced))); err != nil {
		return err
	}

	replaces[origPath] = dst.Name()
	return nil
}

func utf16FromString(s string) ([]uint16, error) {
	if strings.IndexByte(s, 0) != -1 {
		return nil, fmt.Errorf("hitsumabushi: the given string must not include a NUL character")
	}
	// TODO: Use golang.org/x/sys/windows.UTF16FromString, but be careful that this is available only on Windows.
	// See https://github.com/golang/go/issues/59971
	return utf16.Encode([]rune(s + "\x00")), nil
}

func replacementFilePath(fn, pkg, os, file string) (string, error) {
	if os != "linux" {
		return "", fmt.Errorf("hitsumabushi: %s() is not available in this environment: GOOS: %s", fn, os)
	}

	tokens := strings.Split(goVersion(), ".")
	major, err := strconv.Atoi(tokens[0])
	if err != nil {
		return "", err
	}
	minor, err := strconv.Atoi(tokens[1])
	if err != nil {
		return "", err
	}
	if major == 1 && minor < 19 {
		return "", fmt.Errorf("hitsumabushi: %s() is not available in this environment: Go version: %s", fn, runtime.Version())
	}

	dir, err := goPkgDir(pkg, os)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, file), nil
}

// ClockFilePath returns a C file's path for the clock functions.
// This file works only when linux is specified as the GOOS option.
//
// The file includes this function:
//
//   - int hitsumabushi_clock_gettime(clockid_t clk_id, struct timespec *tp)
//
// The default implementation calls clock_gettime.
func ClockFilePath(os string) (string, error) {
	return replacementFilePath("ClockFilePath", "runtime/cgo", os, "hitsumabushi_clock_"+os+".c")
}

// FutexFilePath returns a C file's path for the futex functions.
// This file works only when linux is specified as the GOOS option.
//
// The file includes this function:
//
//   - int32_t hitsumabushi_futex(uint32_t *uaddr, int32_t futex_op, uint32_t val, const struct timespec *timeout, uint32_t *uaddr2, uint32_t val3)
//
// The default implementation is a pseudo futex by pthread.
func FutexFilePath(os string) (string, error) {
	return replacementFilePath("FutexFilePath", "runtime/cgo", os, "hitsumabushi_futex_"+os+".c")
}

// FilesystemFilePath returns a C file's path for the filesystem functions.
// This file works only when linux is specified as the GOOS option.
//
// The file includes these functions:
//
//   - int32_t hitsumabushi_closefd(int32_t fd)
//   - int32_t hitsumabushi_open(char *name, int32_t mode, int32_t perm)
//   - int32_t hitsumabushi_read(int32_t fd, void *p, int32_t n)
//   - int32_t hitsumabushi_write1(uintptr_t fd, void *p, int32_t n)
//   - int32_t hitsumabushi_lseek(uintptr_t fd, off_t offset, int32_t whence)
//   - int32_t hitsumabushi_fcntl(int32_t fd, int32_t cmd, int32_t arg)
//   - int32_t hitsumabushi_fstat(int32_t fd, struct stat *stat)
//   - int32_t hitsumabushi_renameat(int32_t fd1, char* name1, int32_t fd2, char* name2)
//   - int32_t hitsumabushi_fstatat(int32_t fd, char* name, struct stat* p, int32_t flags)
//
// The default implementation only handles stdout, stderr, and some pseudo-files.
func FilesystemFilePath(os string) (string, error) {
	return replacementFilePath("FilesystemFilePath", "runtime/cgo", os, "hitsumabushi_filesystem_"+os+".c")
}

// MemoryFilePath returns a C file's path for the memory functions.
// This file works only when linux is specified as the GOOS option.
//
// The file includes these functions:
//
//   - void* hitsumabushi_sysAllocOS(uintptr_t n)
//   - void hitsumabushi_sysUnusedOS(void* v, uintptr_t n)
//   - void hitsumabushi_sysUsedOS(void* v, uintptr_t n)
//   - void hitsumabushi_sysHugePageOS(void* v, uintptr_t n)
//   - void hitsumabushi_sysFreeOS(void* v, uintptr_t n)
//   - void hitsumabushi_sysFaultOS(void* v, uintptr_t n)
//   - void* hitsumabushi_sysReserveOS(void* v, uintptr_t n)
//   - void hitsumabushi_sysMapOS(void* v, uintptr_t n)
//
// The default implementation is a pseudo allocation by calloc without free.
//
// For the implementation details, see https://cs.opensource.google/go/go/+/master:src/runtime/mem.go .
//
// Note that allocated memory address must be aligned to 1 << 9 bytes as of Go 1.25.
func MemoryFilePath(os string) (string, error) {
	return replacementFilePath("MemoryFilePath", "runtime/cgo", os, "hitsumabushi_mem_"+os+".c")
}

// CPUFilePath returns a C file's path for the CPU functions.
// This file works only when linux is specified as the GOOS option.
//
// The file includes this function:
//
//   - int32_t hitsumabushi_getproccount()
//
// The default implementation uses 1.
func CPUFilePath(os string) (string, error) {
	return replacementFilePath("CPUFilePath", "runtime/cgo", os, "hitsumabushi_cpu_"+os+".c")
}

func ThreadFilePath(os string) (string, error) {
	return replacementFilePath("ThreadFilePath", "runtime/cgo", os, "hitsumabushi_thread_"+os+".c")
}
