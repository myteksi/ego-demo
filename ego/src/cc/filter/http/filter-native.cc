// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "common/common/empty_string.h"

#include "filter.h"

namespace Envoy {
namespace Http {

void GoHttpFilter::setDecoderFilterCallbacks(StreamDecoderFilterCallbacks& callbacks) {
  ASSERT(0 < pins_.load());
  ASSERT(0 == decoderCallbacks_);
  // Assuming dispatcher is shared between callbacks
  // https://github.com/envoyproxy/envoy/blob/master/include/envoy/thread_local/thread_local.h
  ASSERT(nullptr == dispatcher_);

  decoderCallbacks_ = &callbacks;
  dispatcher_ = &callbacks.dispatcher();

  // Handle logic can not create fitler on Go-side
  if (cgoTag_ == 0) {
    decoderCallbacks_->sendLocalReply(Code::InternalServerError, EMPTY_STRING, nullptr,
                                      absl::nullopt, EMPTY_STRING);
  }
}

void GoHttpFilter::setEncoderFilterCallbacks(StreamEncoderFilterCallbacks& callbacks) {
  ASSERT(0 < pins_.load());
  ASSERT(0 == encoderCallbacks_);

  encoderCallbacks_ = &callbacks;
}

uint64_t GoHttpRouteSpecificFilterConfig::cgoTag(std::string filterName) const {
  auto it = filters_.find(filterName);
  if (it != filters_.cend()) {
    return it->second;
  }
  return 0;
}

GoHttpRouteSpecificFilterConfig::~GoHttpRouteSpecificFilterConfig() { onDestroy_(); }

} // namespace Http
} // namespace Envoy