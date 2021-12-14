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
	var options []hitsumabushi.Option
	for _, arg := range os.Args[1:] {
		options = append(options, hitsumabushi.TestPkg(arg))
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
