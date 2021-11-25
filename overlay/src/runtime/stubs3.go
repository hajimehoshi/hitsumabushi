// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !aix && !darwin && !freebsd && !openbsd && !plan9 && !solaris
// +build !aix,!darwin,!freebsd,!openbsd,!plan9,!solaris

package runtime

import (
	"internal/abi"
	"unsafe"
)

//go:nosplit
//go:cgo_unsafe_args
func nanotime1() int64 {
	var ret int64
	var ret2 = &ret
	libcCall(unsafe.Pointer(abi.FuncPCABI0(nanotime1_trampoline)), unsafe.Pointer(&ret2))
	return ret
}
func nanotime1_trampoline()
