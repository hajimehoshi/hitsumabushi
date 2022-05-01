// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 Hajime Hoshi

//go:build ignore
// +build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/hajimehoshi/hitsumabushi"
)

var (
	flagArgs = flag.String("args", "", "arguments")
	flagQEMU = flag.Bool("qemu", false, "use QEMU")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	args, err := splitArgs(*flagArgs)
	if err != nil {
		return err
	}

	path, err := createJSON(args)
	if err != nil {
		return err
	}

	if err := buildTestBinary(path, args); err != nil {
		return err
	}

	if err := runTestBinary(); err != nil {
		return err
	}

	return nil
}

func createJSON(args []string) (string, error) {
	args = append([]string{"hitsumabushi_program"}, args...)
	options := []hitsumabushi.Option{
		hitsumabushi.Args(args...),
		hitsumabushi.TestPkg(args[len(args)-1]),
	}
	overlayJSON, err := hitsumabushi.GenOverlayJSON(options...)
	if err != nil {
		return "", err
	}

	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := f.Write(overlayJSON); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func buildTestBinary(jsonPath string, args []string) error {
	cmd := exec.Command("go", "test", "-c", "-vet=off", "-overlay="+jsonPath, "-o=test")
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = append(os.Environ(),
		"GOOS=linux",
		"GOARCH=arm64",
		"CGO_ENABLED=1",
		"CGO_CFLAGS=-fno-common -fno-short-enums -ffunction-sections -fdata-sections -fPIC -g -O3")
	if *flagQEMU {
		cmd.Env = append(cmd.Env, "CC=aarch64-linux-gnu-gcc")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func runTestBinary() error {
	var cmd *exec.Cmd
	if *flagQEMU {
		cmd = exec.Command("qemu-aarch64", "test")
		cmd.Env = append(os.Environ(), "QEMU_LD_PREFIX=/usr/aarch64-linux-gnu")
	} else {
		cmd = exec.Command("test")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func splitArgs(s string) ([]string, error) {
	// Copied from cmd/internal/quoted/quoted.go

	// Split fields allowing '' or "" around elements.
	// Quotes further inside the string do not count.
	var f []string
	for len(s) > 0 {
		for len(s) > 0 && isSpaceByte(s[0]) {
			s = s[1:]
		}
		if len(s) == 0 {
			break
		}
		// Accepted quoted string. No unescaping inside.
		if s[0] == '"' || s[0] == '\'' {
			quote := s[0]
			s = s[1:]
			i := 0
			for i < len(s) && s[i] != quote {
				i++
			}
			if i >= len(s) {
				return nil, fmt.Errorf("unterminated %c string", quote)
			}
			f = append(f, s[:i])
			s = s[i+1:]
			continue
		}
		i := 0
		for i < len(s) && !isSpaceByte(s[i]) {
			i++
		}
		f = append(f, s[:i])
		s = s[i:]
	}
	return f, nil
}
