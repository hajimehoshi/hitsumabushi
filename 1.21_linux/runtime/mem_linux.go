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
//go:cgo_unsafe_args
func sysAllocOS(n uintptr) (ptr unsafe.Pointer) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sysAllocOS_trampoline)), unsafe.Pointer(&n))
	return
}
func sysAllocOS_trampoline(n uintptr, size uintptr) uintptr

//go:nosplit
//go:cgo_unsafe_args
func sysUnusedOS(v unsafe.Pointer, n uintptr) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sysUnusedOS_trampoline)), unsafe.Pointer(&v))
}
func sysUnusedOS_trampoline(n uintptr, size uintptr)

//go:nosplit
//go:cgo_unsafe_args
func sysUsedOS(v unsafe.Pointer, n uintptr) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sysUsedOS_trampoline)), unsafe.Pointer(&v))
}
func sysUsedOS_trampoline(n uintptr, size uintptr)

//go:nosplit
//go:cgo_unsafe_args
func sysHugePageOS(v unsafe.Pointer, n uintptr) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sysHugePageOS_trampoline)), unsafe.Pointer(&v))
}
func sysHugePageOS_trampoline(n uintptr, size uintptr)

func sysNoHugePageOS(v unsafe.Pointer, n uintptr) {
}

func sysHugePageCollapse(v unsafe.Pointer, n uintptr) {
}

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//
//go:nosplit
//go:cgo_unsafe_args
func sysFreeOS(v unsafe.Pointer, n uintptr) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sysFreeOS_trampoline)), unsafe.Pointer(&v))
}
func sysFreeOS_trampoline(n uintptr, size uintptr)

//go:nosplit
//go:cgo_unsafe_args
func sysFaultOS(v unsafe.Pointer, n uintptr) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sysFaultOS_trampoline)), unsafe.Pointer(&v))
}
func sysFaultOS_trampoline(n uintptr, size uintptr)

//go:nosplit
//go:cgo_unsafe_args
func sysReserveOS(v unsafe.Pointer, n uintptr) (ptr unsafe.Pointer) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sysReserveOS_trampoline)), unsafe.Pointer(&v))
	return
}
func sysReserveOS_trampoline(n uintptr, size uintptr) uintptr

//go:nosplit
//go:cgo_unsafe_args
func sysMapOS(v unsafe.Pointer, n uintptr) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(sysMapOS_trampoline)), unsafe.Pointer(&v))
}
func sysMapOS_trampoline(n uintptr, size uintptr)
