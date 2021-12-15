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
	args := append([]string{"hitsumabushi_program"}, os.Args[1:len(os.Args)]...)
	overlayJSON, err := hitsumabushi.GenOverlayJSON(
		hitsumabushi.TestPkg(args[len(args)-1]),
		hitsumabushi.Args(args...))
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(overlayJSON); err != nil {
		return err
	}
	return nil
}
