// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 Hajime Hoshi

package hitsumabushi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type Option func(*config)

type config struct {
	testPkgs         []string
	numCPU           int
	args             []string
	clockGettimeName string
	futexName        string
}

// TestPkg represents a package for testing.
// When generating a JSON, importing `runtime/cgo` is inserted in the testing package.
func TestPkg(pkg string) Option {
	return func(cfg *config) {
		cfg.testPkgs = append(cfg.testPkgs, pkg)
	}
}

// NumCPU represents a number of CPU.
// The default value is runtime.NumCPU().
func NumCPU(numCPU int) Option {
	return func(cfg *config) {
		cfg.numCPU = numCPU
	}
}

// Args is arguments when executing.
// The first argument must be a program name.
func Args(args ...string) Option {
	return func(cfg *config) {
		cfg.args = append(cfg.args, args...)
	}
}

// ReplaceClockGettime replaces the C function `clock_gettime` with the given name.
// If name is an empty string, the function is not replaced.
// This is useful for special environments where `clock_gettime` doesn't work correctly.
func ReplaceClockGettime(name string) Option {
	return func(cfg *config) {
		cfg.clockGettimeName = name
	}
}

// ReplaceClockGettime replaces the system call `futex` with the given name.
// If name is an empty string, a pseudo futex implementation is used.
// This is useful for special environments where the pseudo `futex` doesn't work correctly.
func ReplaceFutex(name string) Option {
	return func(cfg *config) {
		cfg.futexName = name
	}
}

func currentDir() string {
	_, currentPath, _, _ := runtime.Caller(1)
	return filepath.Dir(currentPath)
}

var reGoVersion = regexp.MustCompile(`^go(\d+\.\d+)(\.\d+)?$`)

// GenOverlayJSON generates a JSON file for go-build's `-overlay` option.
// GenOverlayJSON returns a JSON file content, or an error if generating it fails.
//
// Now the generated JSON works only for Arm64 so far.
func GenOverlayJSON(options ...Option) ([]byte, error) {
	type overlay struct {
		Replace map[string]string
	}

	cfg := config{
		numCPU: runtime.NumCPU(),
	}
	for _, op := range options {
		op(&cfg)
	}

	m := reGoVersion.FindStringSubmatch(runtime.Version())
	dir := filepath.Join(currentDir(), m[1] + "_" + runtime.GOOS)
	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("the Go version %s and GOOS=%s is not supported", runtime.Version(), runtime.GOOS)
		}
		return nil, err
	}
	replaces := map[string]string{}

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, err
	}

	if err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		const (
			modTypeReplace = ".replace"
			modTypeAppend  = ".append"
			modTypePatch   = ".patch"
		)
		modType := modTypeReplace

		origFilename := filepath.Base(path)
		for _, m := range []string{modTypeAppend, modTypePatch} {
			if strings.HasSuffix(origFilename, m) {
				origFilename = origFilename[:len(origFilename)-len(m)]
				modType = m
				break
			}
		}

		ext := filepath.Ext(origFilename)
		if ext != ".go" && ext != ".c" && ext != ".h" && ext != ".s" {
			return nil
		}

		shortPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		pkg := filepath.ToSlash(filepath.Dir(shortPath))
		origDir, err := goPkgDir(pkg)
		if err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		if err := os.MkdirAll(filepath.Join(tmpDir, pkg), 0755); err != nil {
			return err
		}
		dst, err := os.Create(filepath.Join(tmpDir, pkg, origFilename))
		if err != nil {
			return err
		}
		defer dst.Close()

		origPath := filepath.Join(origDir, origFilename)

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

			p, err := parsePatch(shortPath, src)
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

		switch {
		case pkg == "runtime/cgo" && origFilename == "gcc_linux_arm64.c":
			// The number of CPU is defined at runtime/cgo/gcc_linux_arm64.c
			numBytes := (cfg.numCPU-1)/8 + 1
			tmpl := `
int32_t c_sched_getaffinity(pid_t pid, size_t cpusetsize, void *mask) {
{{.Masking}}
  // https://man7.org/linux/man-pages/man2/sched_setaffinity.2.html
  // > On success, the raw sched_getaffinity() system call returns the
  // > number of bytes placed copied into the mask buffer;
  return {{.NumBytes}};
}
`
			n := cfg.numCPU
			var masking string
			for i := 0; i < numBytes; i++ {
				mask := 0
				m := 8
				if n < m {
					m = n
				}
				for j := 0; j < m; j++ {
					mask |= 1 << j
				}
				masking += fmt.Sprintf("  ((char*)mask)[%d] = 0x%x;\n", i, mask)
				n -= 8
			}

			tmpl = strings.ReplaceAll(tmpl, "{{.Masking}}", masking)
			tmpl = strings.ReplaceAll(tmpl, "{{.NumBytes}}", fmt.Sprintf("%d", numBytes))
			if _, err := dst.Write([]byte(tmpl)); err != nil {
				return err
			}
		}

		replaces[origPath] = dst.Name()
		return nil
	}); err != nil {
		return nil, err
	}

	// Add pthread_setaffinity_np.
	{
		indent := "\t\t\t"
		setCPU := []string{
			indent + fmt.Sprintf(`cpu_set_t *cpu_set = CPU_ALLOC(%d);`, cfg.numCPU),
			indent + fmt.Sprintf(`size_t size = CPU_ALLOC_SIZE(%d);`, cfg.numCPU),
			indent + `CPU_ZERO_S(size, cpu_set);`,
			indent + fmt.Sprintf(`for (int i = 0; i < %d; i++) {`, cfg.numCPU),
			indent + `	CPU_SET_S(i, size, cpu_set);`,
			indent + `}`,
			indent + `pthread_setaffinity_np(*thread, size, cpu_set);`,
			indent + `CPU_FREE(cpu_set);`,
		}

		old := `		err = pthread_create(thread, attr, pfn, arg);
		if (err == 0) {
			pthread_detach(*thread);
			return 0;
		}`

		new := strings.Replace(`		err = pthread_create(thread, attr, pfn, arg);
		if (err == 0) {
			pthread_detach(*thread);
{{.SetCPU}}
			return 0;
		}`, "{{.SetCPU}}", strings.Join(setCPU, "\n"), 1)

		if err := replace(tmpDir, replaces, "runtime/cgo", "gcc_libinit.c", old, new); err != nil {
			return nil, err
		}
	}

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
	for i := int32(0); i < %[2]d; i++ {
		argslice[i] = __argv[i]
	}
}`, argvDef, len(cfg.args))
		if err := replace(tmpDir, replaces, "runtime", "runtime1.go", old, new); err != nil {
			return nil, err
		}
	}

	// Replace clock_gettime.
	if cfg.clockGettimeName != "" {
		old := "#define clock_gettime clock_gettime"
		new := fmt.Sprintf(`void %[1]s(clockid_t, struct timespec *);
#define clock_gettime %[1]s`, cfg.clockGettimeName)
		if err := replace(tmpDir, replaces, "runtime/cgo", "gcc_linux_arm64.c", old, new); err != nil {
			return nil, err
		}
	}

	// Replace futex.
	if cfg.futexName != "" {
		old := "#undef user_futex"
		new := fmt.Sprintf(`int32_t %[1]s(uint32_t *uaddr, int32_t futex_op, uint32_t val, const struct timespec *timeout, uint32_t *uaddr2, uint32_t val3);
#define user_futex %[1]s`, cfg.futexName)
		if err := replace(tmpDir, replaces, "runtime/cgo", "gcc_linux_arm64.c", old, new); err != nil {
			return nil, err
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

		if err := os.MkdirAll(filepath.Join(tmpDir, pkg), 0755); err != nil {
			return nil, err
		}
		dst, err := os.Create(filepath.Join(tmpDir, pkg, filepath.Base(origPath)))
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

	return json.Marshal(&overlay{Replace: replaces})
}

func goPkgDir(pkg string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", pkg)
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

func replace(tmpDir string, replaces map[string]string, pkg string, filename string, old, new string) error {
	origDir, err := goPkgDir(pkg)
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

	if err := os.MkdirAll(filepath.Join(tmpDir, pkg), 0755); err != nil {
		return err
	}
	dst, err := os.Create(filepath.Join(tmpDir, pkg, filepath.Base(origPath)))
	if err != nil {
		return err
	}
	defer dst.Close()

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
