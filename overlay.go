// SPDX-License-Identifier: Apache-2.0

package hitsumabushi

import (
	"bytes"
	"encoding/json"
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
	testPkgs []string
	numCPU   int
}

// TestPkg represents a package for testing.
// When generating a JSON, importing `runtime/cgo` is inserted in the testing package.
func TestPkg(pkg string) Option {
	return func(cfg *config) {
		cfg.testPkgs = append(cfg.testPkgs, pkg)
	}
}

// NumCPU represents a number of CPU.
// The default value is 4.
func NumCPU(numCPU int) Option {
	return func(cfg *config) {
		cfg.numCPU = numCPU
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
		numCPU: 4,
	}
	for _, op := range options {
		op(&cfg)
	}

	m := reGoVersion.FindStringSubmatch(runtime.Version())
	dir := filepath.Join(currentDir(), m[1])
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

		if err := os.MkdirAll(filepath.Join(tmpDir, pkg), 0755); err != nil {
			return err
		}
		tmp, err := os.Create(filepath.Join(tmpDir, pkg, origFilename))
		if err != nil {
			return err
		}
		defer tmp.Close()

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		origPath := filepath.Join(origDir, origFilename)

		switch modType {
		case modTypeReplace:
			if _, err := io.Copy(tmp, src); err != nil {
				return err
			}

		case modTypeAppend:
			orig, err := os.Open(origPath)
			if err != nil {
				return err
			}
			defer orig.Close()

			if _, err := io.Copy(tmp, io.MultiReader(orig, src)); err != nil {
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
			if _, err := io.Copy(tmp, patched); err != nil {
				return err
			}

		default:
			return fmt.Errorf("hitsumabushi: unexpected modType: %s", modType)
		}

		// The number of CPU is defined at runtime/cgo/gcc_linux_arm64.c
		if pkg == "runtime/cgo" && origFilename == "gcc_linux_arm64.c" {
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
			if _, err := tmp.Write([]byte(tmpl)); err != nil {
				return err
			}
		}

		replaces[origPath] = tmp.Name()
		return nil
	}); err != nil {
		return nil, err
	}

	for _, pkg := range cfg.testPkgs {
		origPath, err := goTestFile(pkg)
		if err != nil {
			return nil, err
		}

		pkgName, err := goPkgName(pkg)
		if err != nil {
			return nil, err
		}

		if err := os.MkdirAll(filepath.Join(tmpDir, pkg), 0755); err != nil {
			return nil, err
		}
		tmp, err := os.Create(filepath.Join(tmpDir, pkg, filepath.Base(origPath)))
		if err != nil {
			return nil, err
		}
		defer tmp.Close()

		srcContent, err := os.ReadFile(origPath)
		if err != nil {
			return nil, err
		}

		// This assumes that there is an external test package.
		old := "package " + pkgName + "_test"
		new := old + "\n\n" + `import _ "runtime/cgo"`
		replaced := strings.Replace(string(srcContent), old, new, 1)

		if _, err := io.Copy(tmp, bytes.NewReader([]byte(replaced))); err != nil {
			return nil, err
		}

		replaces[origPath] = tmp.Name()
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

func goTestFile(pkg string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}"+string(filepath.Separator)+"{{index .XTestGoFiles 0}}", pkg)
	cmd.Stderr = &buf
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("hitsumabushi: %v\n%s", err, buf.String())
	}
	return strings.TrimSpace(string(out)), nil
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
