// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "common/common/logger.h"

#include "envoy.h"

// Macro ENVOY_LOG_MISC and logger are declare inside Envoy namespace
// So just wrapp it with a function extern "C" below
namespace Envoy {

void Envoy_log_misc(uint32_t level, absl::string_view tag, absl::string_view message) {
  switch (static_cast<spdlog::level::level_enum>(level)) {
  case spdlog::level::trace:
    ENVOY_LOG_MISC(trace, "[ego_http][{}] {}", tag, message);
    return;
  case spdlog::level::debug:
    ENVOY_LOG_MISC(debug, "[ego_http][{}] {}", tag, message);
    return;
  case spdlog::level::info:
    ENVOY_LOG_MISC(info, "[ego_http][{}] {}", tag, message);
    return;
  case spdlog::level::warn:
    ENVOY_LOG_MISC(warn, "[ego_http][{}] {}", tag, message);
    return;
  case spdlog::level::err:
    ENVOY_LOG_MISC(error, "[ego_http][{}] {}", tag, message);
    return;
  case spdlog::level::critical:
    ENVOY_LOG_MISC(critical, "[ego_http][{}] {}", tag, message);
    return;
  case spdlog::level::off:
    return;
  }
  ENVOY_LOG_MISC(warn, "[ego_http][{}] UNDEFINED LOG LEVEL {}: {}", tag, level, message);
}

} // namespace Envoy

void Envoy_log_misc(uint32_t level, GoStr tag, GoStr message) {
  auto c_message = absl::string_view(message.data, message.len);
  auto c_tag = absl::string_view(tag.data, tag.len);
  Envoy::Envoy_log_misc(level, c_tag, c_message);
}
