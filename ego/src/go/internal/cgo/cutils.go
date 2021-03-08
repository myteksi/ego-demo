// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

package main

//#include <string.h>
//#include "ego/src/cc/goc/envoy.h"
import "C"
import (
	"errors"
	"math"
	"reflect"
	"unsafe"

	"github.com/grab/ego/ego/src/go/volatile"
)

// GoStr create an unsafe reference to a Go string for passing it down to C++
//
func GoStr(s string) C.GoStr {
	h := (*reflect.StringHeader)(unsafe.Pointer(&s))
	return C.GoStr{
		len:  C.ulong(h.Len),
		data: (*C.char)(unsafe.Pointer(h.Data)),
	}
}

// GoBuf create an unsafe reference to a Go byte array for passing it down to C++
//
func GoBuf(b []byte) C.GoBuf {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	return C.GoBuf{
		len:  C.ulong(h.Len),
		cap:  C.ulong(h.Cap),
		data: unsafe.Pointer(h.Data),
	}
}

// We have agreed on a sweet little convention for returning errors
// from C++ land: static, constant C strings. CErr wraps them into a
// Go error.
//
func CErr(ptr *C.char) error {
	if nil == ptr {
		return nil
	}
	return errors.New(string(CStr(ptr)))
}

// CStr is very dangerous, because if the returned string slice escapes,
// its contents may be modified later. If you are unsure, do use C.GoString.
//
func CStr(ptr *C.char) volatile.String {

	// prevent null pointer access
	if nil == ptr {
		return ""
	}

	return CStrN(ptr, C.strlen(ptr))
}

// CStrN is very dangerous, because if the returned string slice escapes,
// its contents may be modified later. If you are unsure, do use C.GoStringN.
//
func CStrN(ptr *C.char, n C.size_t) volatile.String {

	// prevent null pointer access
	if nil == ptr {
		if 0 != n {
			panic("C string is null")
		}
		return ""
	}

	// ensure no loss occurs during conversion.
	if n != C.size_t(int(n)) || int(n) < 0 {
		panic("C string too long")
	}
	h := reflect.StringHeader{uintptr(unsafe.Pointer(ptr)), int(n)}
	return *(*volatile.String)(unsafe.Pointer(&h))
}

// CBytes is very dangerous, because if the returned byte slice escapes,
// its contents may be modified later. If you are unsure, do use C.GoBytes.
//
func CBytes(ptr unsafe.Pointer, size, cap C.size_t) volatile.Bytes {

	// The maximum address space can be larger than 4GB. Apparently, 50bits are
	// supported for 64bit architectures, so we bump up the value in that case.
	const maxCap = int(math.MaxInt32) | int((^uint(0))>>14)

	// ensure no loss occurs during conversion and capacity isn't too large
	if C.size_t(int(cap)) != cap || int(cap) < 0 || maxCap < int(cap) {
		panic("C buffer too large")
	}
	if C.size_t(int(size)) != size || int(size) < 0 || int(cap) < int(size) {
		panic("Invalid C buffer")
	}
	// https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
	return volatile.Bytes((*[maxCap]byte)(ptr)[:size:int(cap)])
}

func CLong(val C.int64_t) int64 {
	return int64(val)
}

func GoBool(val bool) C.int {
	if val {
		return C.int(1)
	}
	return C.int(0)
}
