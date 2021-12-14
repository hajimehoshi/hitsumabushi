// SPDX-License-Identifier: Apache-2.0

//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/hitsumabushi"
)

func main() {
	if err := build(); err != nil {
		log.Fatal(err)
	}
}

func build() error {
	overlayJSON, err := hitsumabushi.GenOverlayJSON()
	if err != nil {
		return err
	}
	fmt.Println(overlayJSON)
	return nil
}
