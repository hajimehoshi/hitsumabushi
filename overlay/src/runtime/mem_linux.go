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

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//go:nosplit
func sysAlloc(n uintptr, sysStat *sysMemStat) unsafe.Pointer {
	p := sysReserve(nil, n)
	sysMap(p, n, sysStat)
	return p
}

func sysUnused(v unsafe.Pointer, n uintptr) {
}

func sysUsed(v unsafe.Pointer, n uintptr) {
}

func sysHugePage(v unsafe.Pointer, n uintptr) {
}

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//go:nosplit
func sysFree(v unsafe.Pointer, n uintptr, sysStat *sysMemStat) {
	sysStat.add(-int64(n))
}

func sysFault(v unsafe.Pointer, n uintptr) {
}

func sysReserve(v unsafe.Pointer, n uintptr) unsafe.Pointer {
	if v != nil {
		return nil
	}
	var ptr unsafe.Pointer
	calloc(&ptr, n, 1)
	return ptr
}

//go:nosplit
//go:cgo_unsafe_args
func calloc(ptr *unsafe.Pointer, n uintptr, size uintptr) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(calloc_trampoline)), unsafe.Pointer(&ptr))
}
func calloc_trampoline()

func sysMap(v unsafe.Pointer, n uintptr, sysStat *sysMemStat) {
	sysStat.add(int64(n))
}
