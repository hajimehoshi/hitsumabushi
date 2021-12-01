// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !aix && !darwin && !js && !openbsd && !plan9 && !solaris && !windows
// +build !aix,!darwin,!js,!openbsd,!plan9,!solaris,!windows

package runtime

import (
	"internal/abi"
	"unsafe"
)

// read calls the read system call.
// It returns a non-negative number of bytes written or a negative errno value.
//go:nosplit
//go:cgo_unsafe_args
func read(fd int32, p unsafe.Pointer, n int32) int32 {
	ret := libcCall(unsafe.Pointer(abi.FuncPCABI0(read_trampoline)), unsafe.Pointer(&fd))
	KeepAlive(p)
	return ret
}
func read_trampoline()

//go:nosplit
//go:cgo_unsafe_args
func closefd(fd int32) int32 {
	return libcCall(unsafe.Pointer(abi.FuncPCABI0(closefd_trampoline)), unsafe.Pointer(&fd))
}
func closefd_trampoline()

func exit(code int32)

//go:nosplit
//go:cgo_unsafe_args
func usleep(usec uint32) {
	libcCall(unsafe.Pointer(abi.FuncPCABI0(usleep_trampoline)), unsafe.Pointer(&usec))
}
func usleep_trampoline()

//go:nosplit
func usleep_no_g(usec uint32) {
	usleep(usec)
}

// write calls the write system call.
// It returns a non-negative number of bytes written or a negative errno value.
//go:nosplit
//go:cgo_unsafe_args
func write1(fd uintptr, p unsafe.Pointer, n int32) int32 {
	ret := libcCall(unsafe.Pointer(abi.FuncPCABI0(write1_trampoline)), unsafe.Pointer(&fd))
	KeepAlive(p)
	return ret
}
func write1_trampoline()

//go:nosplit
//go:cgo_unsafe_args
func open(name *byte, mode, perm int32) int32 {
	ret := libcCall(unsafe.Pointer(abi.FuncPCABI0(open_trampoline)), unsafe.Pointer(&name))
	KeepAlive(name)
	return ret
}
func open_trampoline()

// return value is only set on linux to be used in osinit()
func madvise(addr unsafe.Pointer, n uintptr, flags int32) int32

// exitThread terminates the current thread, writing *wait = 0 when
// the stack is safe to reclaim.
//
//go:noescape
func exitThread(wait *uint32)

//go:linkname c_calloc c_calloc
//go:cgo_import_static c_calloc
var c_calloc byte

//go:linkname c_closefd c_closefd
//go:cgo_import_static c_closefd
var c_closefd byte

//go:linkname c_gettid c_gettid
//go:cgo_import_static c_gettid
var c_gettid byte

//go:linkname c_nanotime1 c_nanotime1
//go:cgo_import_static c_nanotime1
var c_nanotime1 byte

//go:linkname c_open c_open
//go:cgo_import_static c_open
var c_open byte

//go:linkname c_osyield c_osyield
//go:cgo_import_static c_osyield
var c_osyield byte

//go:linkname c_read c_read
//go:cgo_import_static c_read
var c_read byte

//go:linkname c_sched_getaffinity c_sched_getaffinity
//go:cgo_import_static c_sched_getaffinity
var c_sched_getaffinity byte

//go:linkname c_usleep c_usleep
//go:cgo_import_static c_usleep
var c_usleep byte

//go:linkname c_write1 c_write1
//go:cgo_import_static c_write1
var c_write1 byte
