// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include <cstdint>

#include "envoy/buffer/buffer.h"

#include "envoy.h"

// BufferInstance
uint64_t BufferInstance_copyOut(void* bufferInstance, size_t start, GoBuf buf) {
  auto that = static_cast<Envoy::Buffer::Instance*>(bufferInstance);
  if (that->length() <= start)
    return 0;

  auto size = that->length() - start;
  if (buf.len < size)
    size = buf.len;

  that->copyOut(start, size, buf.data);
  return size;
}

uint64_t BufferInstance_length(void* bufferInstance) {
  auto that = static_cast<Envoy::Buffer::Instance*>(bufferInstance);
  return that->length();
}

uint64_t BufferInstance_getRawSlicesCount(void* bufferInstance) {
  auto that = static_cast<Envoy::Buffer::Instance*>(bufferInstance);
  return that->getRawSlices().size();
}

uint64_t BufferInstance_getRawSlices(void* bufferInstance, uint64_t max, GoBuf* dest) {
  auto that = static_cast<Envoy::Buffer::Instance*>(bufferInstance);

  auto len = 0;
  for (const auto& slice : that->getRawSlices(max)) {
    dest->data = slice.mem_;
    dest->len = dest->cap = slice.len_;
    dest++;
    len++;
  }
  return len;
}
