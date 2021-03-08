// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package volatile

import "C"
import (
	"reflect"
	"unsafe"
)

// String should not be kept anywhere beyond the life cycle of the C callback.
// In order to secure its contents, use String.Copy()
type String string

// Bytes should not be kept anywhere beyond the life cycle of the C callback.
// In order to secure its contents, use Bytes.Copy()
type Bytes []byte

// Copy creates a non-volatile, "safe" copy of the volatile string
func (s String) Copy() string {
	h := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return C.GoStringN((*C.char)(unsafe.Pointer(h.Data)), C.int(h.Len))
}

// Copy creates a non-volatile, "safe" copy of the volatile buffer
func (b Bytes) Copy() []byte {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	return C.GoBytes(unsafe.Pointer(h.Data), C.int(h.Cap))[:h.Len]
}
