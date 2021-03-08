#pragma once
// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include <cstdint>
#include "envoy/server/filter_config.h"

namespace Envoy {
namespace Http {

  struct SpanGroupConstantValues {
    const int EncoderActiveSpan = 0;
    const int DecoderActiveSpan = -1;
  };

  using SpanGroupConstants = ConstSingleton<SpanGroupConstantValues>;

  class SpanGroup {
    public:
      SpanGroup();
      intptr_t spawnChildSpan(const intptr_t parent_span_id, std::string &name);
      Envoy::Tracing::Span& getSpan(const intptr_t span_id);
      void removeSpan(const intptr_t span_id);
      void setDecoderFilterCallbacks(StreamDecoderFilterCallbacks& callbacks);
      void setEncoderFilterCallbacks(StreamEncoderFilterCallbacks& callbacks);

    private:
      void setSpan(const intptr_t id, Envoy::Tracing::SpanPtr span);
      Envoy::Tracing::Span& getSpan_(const intptr_t span_id);
      
      std::map<intptr_t, std::unique_ptr<Envoy::Tracing::Span>> spans_;
      std::mutex spans_mutex;
      // Do only access from dispatcher context. Do check if non-0 before use.
      StreamDecoderFilterCallbacks* decoderCallbacks_;
      StreamEncoderFilterCallbacks* encoderCallbacks_;
  };

  using SpanGroupPtr = std::unique_ptr<SpanGroup>;

}
}