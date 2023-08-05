// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 Hajime Hoshi

//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/hitsumabushi"
)

func main() {
	if err := build(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func build() error {
	options := []hitsumabushi.Option{
		hitsumabushi.Args(os.Args...),
		hitsumabushi.GOOS("linux"),
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
