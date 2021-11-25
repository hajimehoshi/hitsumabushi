package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/hajimehoshi/gooslibc/overlay"
)

var (
	flagO = flag.String("o", "", "output")
)

func main() {
	flag.Parse()
	if err := build(); err != nil {
		log.Fatal(err)
	}
}

func build() error {
	if *flagO == "" {
		return fmt.Errorf("-o must be speicified")
	}

	overlayJSON, err := overlay.GenOverlayJSON()
	if err != nil {
		return err
	}

	// c-archive
	args := []string{
		"build",
		"-buildmode=c-archive",
		"-o=" + *flagO,
		"-overlay=" + overlayJSON,
	}
	args = append(args, flag.Args()...)
	cmd := exec.Command("go", args...)
	cflags := `-fno-common -fno-short-enums -ffunction-sections -fdata-sections -fPIC -g -O3`
	cmd.Env = append(os.Environ(),
		`GOOS=linux`,
		`GOARCH=arm64`,
		`CGO_ENABLED=1`,
		`CGO_CFLAGS=`+cflags)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
