// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "common/common/lock_guard.h"
#include "common/http/utility.h"

#include "filter.h"

namespace Envoy {
namespace Http {

void GoHttpFilter::pin() {
  ASSERT(0 < pins_.load());

  pins_++;
}

void GoHttpFilter::unpin() {
  ASSERT(0 < pins_.load());

  if (0 == --pins_) {

    // If we got here, onDestroy() must already be waiting for this!

    // "Even if the shared variable is atomic, it must be modified under the
    // mutex in order to correctly publish the modification to the waiting
    // thread." [https://en.cppreference.com/w/cpp/thread/condition_variable]
    Envoy::Thread::LockGuard lk_m(m_);

    // notify onDestroy()
    cv_.notifyOne();
  }
}

void GoHttpFilter::post(uint64_t tag) {
  ASSERT(0 < pins_.load());

  // ref() and dispatcher_ are  guarded by 0 < pins_
  dispatcher_->post([this, tag, keepalive = ref()]() { onPost(tag); });
}

void GoHttpFilter::log(uint32_t level, absl::string_view message) {
  switch (static_cast<spdlog::level::level_enum>(level)) {
  case spdlog::level::trace:
    ENVOY_LOG(trace, "[ego_http][{}] [{}] {}", config_->filter(), x_request_id_, message);
    return;
  case spdlog::level::debug:
    ENVOY_LOG(debug, "[ego_http][{}] [{}] {}", config_->filter(), x_request_id_, message);
    return;
  case spdlog::level::info:
    ENVOY_LOG(info, "[ego_http][{}] [{}] {}", config_->filter(), x_request_id_, message);
    return;
  case spdlog::level::warn:
    ENVOY_LOG(warn, "[ego_http][{}] [{}] {}", config_->filter(), x_request_id_, message);
    return;
  case spdlog::level::err:
    ENVOY_LOG(error, "[ego_http][{}] [{}] {}", config_->filter(), x_request_id_, message);
    return;
  case spdlog::level::critical:
    ENVOY_LOG(critical, "[ego_http][{}] [{}] {}", config_->filter(), x_request_id_, message);
    return;
  case spdlog::level::off:
    return;
  }
  ENVOY_LOG(warn, "[ego_http][{}] [{}] UNDEFINED LOG LEVEL {}: {}", config_->filter(),
            x_request_id_, level, message);
}

StreamDecoderFilterCallbacks* GoHttpFilter::decoderCallbacks() {
  ASSERT(0 != decoderCallbacks_);
  return decoderCallbacks_;
}

StreamEncoderFilterCallbacks* GoHttpFilter::encoderCallbacks() {
  ASSERT(0 != encoderCallbacks_);
  return encoderCallbacks_;
}

StreamFilterCallbacks* GoHttpFilter::streamFilterCallbacks(int encoder) {
  if(encoder != 0) {
    return encoderCallbacks();
  }
  return decoderCallbacks();
}

Secret::GenericSecretConfigProviderSharedPtr GoHttpFilter::genericSecretConfigProvider() {
  return secret_provider_;
}

Api::Api& GoHttpFilter::api() { return api_; }

uint64_t GoHttpFilter::resolveMostSpecificPerGoFilterConfigTag() {
  if (decoderCallbacks_ == nullptr) {
    return 0;
  }

  // decoderCallbacks_ is set when requests start.
  // and it's safe to use until onDestroy is called based on envoy/include/envoy/http/filter.h
  // Therefore, resolveMostSpecificPerGoFilterConfigTag can be called in Encode* method.
  auto route = decoderCallbacks_->route();
  if (route == nullptr || route->routeEntry() == nullptr) {
    return 0;
  }

  const auto* config =
      Http::Utility::resolveMostSpecificPerFilterConfig<GoHttpRouteSpecificFilterConfig>(
          GoHttpConstants::get().FilterName, route);

  if (config == nullptr) {
    return 0;
  }

  return config->cgoTag(config_->filter());
}


intptr_t GoHttpFilter::spawnChildSpan(const intptr_t parent_span_id, std::string &name) {
  return span_group_->spawnChildSpan(parent_span_id, name);
}

Envoy::Tracing::Span& GoHttpFilter::getSpan(const intptr_t span_id) {
  return span_group_->getSpan(span_id);
}

void GoHttpFilter::finishSpan(const intptr_t span_id) {
  span_group_->getSpan(span_id).finishSpan();
  span_group_->removeSpan(span_id);
}


} // namespace Http
} // namespace Envoy