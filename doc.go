// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021 Hajime Hoshi

// Package hitsumabushi provides APIs to generate JSON for go-build's `-overlay` option.
// Hitsumabushi aims to make Go programs work on almost everywhere by overwriting system calls with C function calls.
// Now the generated JSON works only for Linux/Amd64, Linux/Arm64, and Windows/Amd64 so far.
// For GOOS=windows, Hitsumabushi replaces some functions that don't work on some special Windows-like systems.
package hitsumabushi
