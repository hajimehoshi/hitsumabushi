package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/hajimehoshi/hitsumabushi"
)

var (
	flagO    = flag.String("o", "", "output")
	flagTest = flag.Bool("test", false, "run test")
	flagV    = flag.Bool("v", false, "-v")
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

	overlayJSON, err := hitsumabushi.GenOverlayJSON()
	if err != nil {
		return err
	}

	// c-archive
	var args []string
	if *flagTest {
		args = []string{
			"test",
			"-c",
			"-vet=off", // See golang/go#44957, golang/go#50044
		}
	} else {
		args = []string{
			"build",
			"-buildmode=c-archive",
		}
	}
	args = append(args, "-overlay="+overlayJSON, "-o="+*flagO)
	if *flagV {
		args = append(args, "-v")
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
