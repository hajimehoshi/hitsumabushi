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

func currentDir() string {
	_, currentPath, _, _ := runtime.Caller(1)
	return filepath.Dir(currentPath)
}

type overlay struct {
	Replace map[string]string
}

var reGoVersion = regexp.MustCompile(`^go(\d+\.\d+)(\.\d+)?$`)

func GenOverlayJSON() (string, error) {
	m := reGoVersion.FindStringSubmatch(runtime.Version())
	dir := filepath.Join(currentDir(), m[1])
	replaces := map[string]string{}

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
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

			p, err := parsePatch(src)
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
			return fmt.Errorf("unexpected modType: %s", modType)
		}

		replaces[origPath] = tmp.Name()
		return nil
	}); err != nil {
		return "", err
	}

	f, err := os.CreateTemp("", "overlay.json")
	if err != nil {
		return "", err
	}
	defer f.Close()

	e := json.NewEncoder(f)
	if err := e.Encode(&overlay{Replace: replaces}); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func goPkgDir(pkg string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", pkg)
	cmd.Stderr = &buf
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%v\n%s", err, buf.String())
	}
	return strings.TrimSpace(string(out)), nil
}
