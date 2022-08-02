// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This implementation is the most fundamental and minimal allocations like Wasm.
// Malloced memory regions are never freed.

package runtime

import (
	"internal/abi"
	"unsafe"
)

// Don't split the stack as this method may be invoked without a valid G, which
// prevents us from allocating more stack.
//
//go:nosplit
func sysAllocOS(n uintptr) unsafe.Pointer {
	return sysReserve(nil, n)
}

func sysUnusedOS(v unsafe.Pointer, n uintptr) {
}

func sysUsedOS(v unsafe.Pointer, n uintptr) {
}

func sysHugePageOS(v unsafe.Pointer, n uintptr) {
}

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//
//go:nosplit
func sysFreeOS(v unsafe.Pointer, n uintptr) {
}

func sysFaultOS(v unsafe.Pointer, n uintptr) {
}

func sysReserveOS(v unsafe.Pointer, n uintptr) unsafe.Pointer {
	if v != nil {
		return nil
	}
	ptr := calloc(n, 1)
	return unsafe.Pointer(ptr)
}

//go:nosplit
//go:cgo_unsafe_args
func calloc(n uintptr, size uintptr) (ret uintptr) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(calloc_trampoline)), unsafe.Pointer(&n))
	return
}
func calloc_trampoline(n uintptr, size uintptr) uintptr

func sysMapOS(v unsafe.Pointer, n uintptr) {
}
