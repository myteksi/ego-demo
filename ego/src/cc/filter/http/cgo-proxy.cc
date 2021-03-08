// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "cgo-proxy.h"

#include "ego/src/cc/goc/goc.h"
#include "ego/src/go/internal/cgo/cgo.h"

namespace Envoy {
namespace Http {

CgoProxyImpl::CgoProxyImpl() = default;

CgoProxyImpl::~CgoProxyImpl() = default;

unsigned long long CgoProxyImpl::GoHttpFilterCreate(void* native, unsigned long long factory_tag,
                                                    unsigned long long filter_slot) {
  return Cgo_GoHttpFilter_Create(native, factory_tag, filter_slot);
}

void CgoProxyImpl::GoHttpFilterOnDestroy(unsigned long long filter_tag) {
  Cgo_GoHttpFilter_OnDestroy(filter_tag);
}

long long CgoProxyImpl::GoHttpFilterDecodeHeaders(unsigned long long filter_tag, void* headers,
                                                  int end_stream) {
  return Cgo_GoHttpFilter_DecodeHeaders(filter_tag, headers, end_stream);
}

long long CgoProxyImpl::GoHttpFilterDecodeData(unsigned long long filter_tag, void* buffer,
                                               int end_stream) {
  return Cgo_GoHttpFilter_DecodeData(filter_tag, buffer, end_stream);
}

long long CgoProxyImpl::GoHttpFilterDecodeTrailers(unsigned long long filter_tag, void* trailers) {
  return Cgo_GoHttpFilter_DecodeTrailers(filter_tag, trailers);
}

long long CgoProxyImpl::GoHttpFilterEncodeHeaders(unsigned long long filter_tag, void* headers,
                                                  int end_stream) {
  return Cgo_GoHttpFilter_EncodeHeaders(filter_tag, headers, end_stream);
}

long long CgoProxyImpl::GoHttpFilterEncodeData(unsigned long long filter_tag, void* buffer,
                                               int end_stream) {
  return Cgo_GoHttpFilter_EncodeData(filter_tag, buffer, end_stream);
}

void CgoProxyImpl::GoHttpFilterOnPost(unsigned long long filter_tag, unsigned long long post_tag) {
  return Cgo_GoHttpFilter_OnPost(filter_tag, post_tag);
}
} // namespace Http
} // namespace Envoy