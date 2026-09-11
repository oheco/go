// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgo

import _ "unsafe"

//go:cgo_import_static x_cgo_ohos_load_g
//go:linkname x_cgo_ohos_load_g x_cgo_ohos_load_g
//go:linkname _cgo_ohos_load_g _cgo_ohos_load_g
var x_cgo_ohos_load_g byte
var _cgo_ohos_load_g = &x_cgo_ohos_load_g

//go:cgo_import_static x_cgo_ohos_save_g
//go:linkname x_cgo_ohos_save_g x_cgo_ohos_save_g
//go:linkname _cgo_ohos_save_g _cgo_ohos_save_g
var x_cgo_ohos_save_g byte
var _cgo_ohos_save_g = &x_cgo_ohos_save_g

//go:cgo_import_static x_cgo_ohos_environ
//go:linkname x_cgo_ohos_environ x_cgo_ohos_environ
//go:linkname _cgo_ohos_environ _cgo_ohos_environ
var x_cgo_ohos_environ byte
var _cgo_ohos_environ = &x_cgo_ohos_environ
