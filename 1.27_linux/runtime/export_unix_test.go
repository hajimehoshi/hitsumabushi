// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package runtime

const (
	O_WRONLY = _O_WRONLY
	O_CREAT  = _O_CREAT
	O_TRUNC  = _O_TRUNC
)
