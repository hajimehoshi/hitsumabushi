// SPDX-License-Identifier: Apache-2.0

//go:build ignore
// +build ignore

package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/hitsumabushi"
)

func main() {
	if err := build(); err != nil {
		log.Fatal(err)
	}
}

func build() error {
	test := os.Args[1] == "test"
	args := append([]string{"hitsumabushi_program"}, os.Args[2:len(os.Args)]...)
	options := []hitsumabushi.Option{hitsumabushi.Args(args...)}
	if test {
		options = append(options, hitsumabushi.TestPkg(args[len(args)-1]))
	}
	overlayJSON, err := hitsumabushi.GenOverlayJSON(options...)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(overlayJSON); err != nil {
		return err
	}
	return nil
}
