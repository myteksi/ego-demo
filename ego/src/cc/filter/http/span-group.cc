// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "span-group.h"
#include "common/tracing/http_tracer_impl.h"

namespace Envoy {
namespace Http {

  SpanGroup::SpanGroup():decoderCallbacks_(nullptr),encoderCallbacks_(nullptr){}

  intptr_t SpanGroup::spawnChildSpan(const intptr_t parent_span_id, std::string &name) {
    auto span = getSpan(parent_span_id).spawnChild(Envoy::Tracing::EgressConfig::get(), name, decoderCallbacks_->dispatcher().timeSource().systemTime());
    auto id = reinterpret_cast<intptr_t>(span.get());
    setSpan(id, std::move(span));
    return id;
  }

  void SpanGroup::setSpan(const intptr_t id, Envoy::Tracing::SpanPtr span){
    std::lock_guard<std::mutex> guard(spans_mutex);
    spans_[id] = std::move(span);
  }

  Envoy::Tracing::Span& SpanGroup::getSpan_(const intptr_t span_id){
    std::lock_guard<std::mutex> guard(spans_mutex);
    auto search = spans_.find(span_id);

    if (search != spans_.end()) {
      return *search->second;
    }

    return Envoy::Tracing::NullSpan::instance();
  }

  void SpanGroup::removeSpan(const intptr_t span_id){
    std::lock_guard<std::mutex> guard(spans_mutex);
    spans_.erase(span_id);
  }

  Envoy::Tracing::Span& SpanGroup::getSpan(const intptr_t span_id) {
    if (SpanGroupConstants::get().DecoderActiveSpan == span_id) {
      if (nullptr == decoderCallbacks_) {
        return Envoy::Tracing::NullSpan::instance();
      }
      return decoderCallbacks_->activeSpan();
    }

    if (SpanGroupConstants::get().EncoderActiveSpan == span_id) {
      if (nullptr == encoderCallbacks_) {
        return Envoy::Tracing::NullSpan::instance();
      }
      return encoderCallbacks_->activeSpan();
    }

    return getSpan_(span_id);
  }

  void SpanGroup::setDecoderFilterCallbacks(StreamDecoderFilterCallbacks& callbacks) {
    decoderCallbacks_ = &callbacks;
  }

  void SpanGroup::setEncoderFilterCallbacks(StreamEncoderFilterCallbacks& callbacks) {
    encoderCallbacks_ = &callbacks;
  }

}
}