// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file


// This file contains upcall proxies to the Go filter implementation. Please
// avoid having too much custom logic in here: the intent is to handle as much
// of the filter logic as possible in Go!

#include "common/common/lock_guard.h"

#include "ego/src/cc/goc/goc.h"
#include "ego/src/go/internal/cgo/cgo.h"
#include "filter.h"

namespace Envoy {
namespace Http {

static thread_local struct CgoHttpFilterFactorySlot_ {
  uint64_t value;
  ~CgoHttpFilterFactorySlot_() { Cgo_ReleaseHttpFilterFactorySlot(value); }
} cgoHttpFilterFactorySlot{Cgo_AcquireHttpFilterFactorySlot()};

GoHttpFilterConfig::GoHttpFilterConfig(const ego::http::Settings& proto,
                                       const Stats::ScopeSharedPtr& scope)
    : cgoSlot_(cgoHttpFilterFactorySlot.value), filter_(proto.filter()), scope_(scope) {
  auto filter = proto.filter();
  auto settings = proto.settings().value();
  cgoTag_ = Cgo_GoHttpFilterFactory_Create(cgoSlot_, const_cast<char*>(filter.c_str()),
                                           filter.size(), const_cast<char*>(settings.c_str()),
                                           settings.size(), scope.get());

  if (cgoTag_ == 0) {
    auto errMsg = std::string(
        "invoke Cgo_GoHttpFilterFactory_Create failed, check error from factory for detail");
    if (proto.crash_on_errors()) {
      throw EnvoyException(errMsg);
    } else {
      ENVOY_LOG(error, "[ego_http][{}] {}", proto.filter(), errMsg);
    }
  }
}

inline bool GoHttpFilterConfig::cgoSafe() { return cgoSlot_ == cgoHttpFilterFactorySlot.value; }

GoHttpFilterConfig::~GoHttpFilterConfig() {
  ASSERT(cgoSafe());
  Cgo_GoHttpFilterFactory_OnDestroy(cgoTag_);
}

static thread_local struct CgoHttpRouteSpecificFilterConfigSlot_ {
  uint64_t value;
  ~CgoHttpRouteSpecificFilterConfigSlot_() { Cgo_ReleaseRouteSpecificFilterConfigSlot(value); }
} cgoHttpRouteSpecificFilterConfigSlot_{Cgo_AcquireRouteSpecificFilterConfigSlot()};

GoHttpRouteSpecificFilterConfig::GoHttpRouteSpecificFilterConfig(
    const ego::http::SettingsPerRoute& proto)
    : cgoSlot_(cgoHttpRouteSpecificFilterConfigSlot_.value) {

  // onDestroy_ is called in destructor to clean up resources at Go side.
  // this is a work-around for issue that destructor of RouteSpecificFilterConfig is a virtual
  // functions. Because of that, the destructor can't be implemented in this file.
  onDestroy_ = [this]() {
    ASSERT(this->cgoSafe());
    for (const auto& it : this->filters_) {
      Cgo_RouteSpecificFilterConfig_dtor(it.second);
    }
  };
  for (const auto& it : proto.filters()) {
    auto filterName = it.first;
    auto filterConfig = it.second;
    auto cgoTag = Cgo_RouteSpecificFilterConfig_Create(
        cgoSlot_, const_cast<char*>(filterName.c_str()), filterName.size(),
        const_cast<char*>(filterConfig.value().c_str()), filterConfig.value().size());
    // handle invalid filter name or configuration
    ASSERT(0 != cgoTag);

    filters_.insert(std::pair<std::string, uint64_t>(filterName, cgoTag));
  }
}

inline bool GoHttpRouteSpecificFilterConfig::cgoSafe() {
  return cgoSlot_ == cgoHttpRouteSpecificFilterConfigSlot_.value;
}

static thread_local struct CgoHttpFilterSlot_ {
  uint64_t value;
  ~CgoHttpFilterSlot_() { Cgo_ReleaseHttpFilterSlot(value); }
} cgoHttpFilterSlot{Cgo_AcquireHttpFilterSlot()};

GoHttpFilter::GoHttpFilter(std::shared_ptr<GoHttpFilterConfig> config, Api::Api& api,
                           Secret::GenericSecretConfigProviderSharedPtr secret_provider,
                           CgoProxyPtr cgo_proxy, SpanGroupPtr span_group)
    : config_(config), decoderCallbacks_(0), encoderCallbacks_(0), dispatcher_(0), pins_(1),
      self_(this), api_(api), secret_provider_(secret_provider), cgo_proxy_(cgo_proxy), span_group_(std::move(span_group)) {
  cgoSlot_ = cgoHttpFilterSlot.value;
  cgoTag_ = cgo_proxy_->GoHttpFilterCreate(this, config->cgoTag_, cgoSlot_);
  // cgoTag_ == 0 means can not create a instance of filter on Go-side
  // we should sendLocalReply here but we don't have decodeCallbacks now
  // Let handle it on setDecoderFilterCallbacks
};

inline bool GoHttpFilter::cgoSafe() {
  return (cgoSlot_ == cgoHttpFilterSlot.value) && dispatcher_ && decoderCallbacks_ &&
         encoderCallbacks_;
}

FilterHeadersStatus GoHttpFilter::decodeHeaders(RequestHeaderMap& headers, bool end_stream) {
  ASSERT(cgoSafe());
  // Get x-request-id for logging
  if (headers.RequestId() != nullptr && x_request_id_ == "") {
    x_request_id_ = headers.RequestId()->value().getStringView();
  }

  return Goc_FilterHeadersStatus(
      cgo_proxy_->GoHttpFilterDecodeHeaders(cgoTag_, &headers, end_stream ? 1 : 0));
}

FilterDataStatus GoHttpFilter::decodeData(Buffer::Instance& buffer, bool end_stream) {
  ASSERT(cgoSafe());
  return Goc_FilterDataStatus(
      cgo_proxy_->GoHttpFilterDecodeData(cgoTag_, &buffer, end_stream ? 1 : 0));
}

FilterTrailersStatus GoHttpFilter::decodeTrailers(RequestTrailerMap& trailers) {
  ASSERT(cgoSafe());
  return Goc_FilterTrailersStatus(cgo_proxy_->GoHttpFilterDecodeTrailers(cgoTag_, &trailers));
}

FilterHeadersStatus GoHttpFilter::encode100ContinueHeaders(ResponseHeaderMap&) {
  // TODO: call to Go
  return Envoy::Http::FilterHeadersStatus::Continue;
}

FilterHeadersStatus GoHttpFilter::encodeHeaders(ResponseHeaderMap& headers, bool end_stream) {
  ASSERT(cgoSafe());
  return Goc_FilterHeadersStatus(
      cgo_proxy_->GoHttpFilterEncodeHeaders(cgoTag_, &headers, end_stream ? 1 : 0));
}

FilterDataStatus GoHttpFilter::encodeData(Buffer::Instance& buffer, bool end_stream) {
  ASSERT(cgoSafe());

  return Goc_FilterDataStatus(
      cgo_proxy_->GoHttpFilterEncodeData(cgoTag_, &buffer, end_stream ? 1 : 0));
}

FilterTrailersStatus GoHttpFilter::encodeTrailers(ResponseTrailerMap&) {
  // TODO: call to Go
  return Envoy::Http::FilterTrailersStatus::Continue;
}

FilterMetadataStatus GoHttpFilter::encodeMetadata(MetadataMap&) {
  // TODO: call to Go
  return Envoy::Http::FilterMetadataStatus::Continue;
}

void GoHttpFilter::encodeComplete() {
  // TODO: call to Go
}

void GoHttpFilter::onDestroy() {
  ASSERT(cgoSafe());

  // best effort to terminate go-routines and other asynchronous activities
  cgo_proxy_->GoHttpFilterOnDestroy(cgoTag_);
  cgoTag_ = 0;
  cgoSlot_ = 0;

  ASSERT(0 < pins_.load());
  if (0 < --pins_) {
    // We're not the last ones to go, so we need to wait until all activity on
    // this filter has ended. This important to implement the contract for
    // StreamDecoderFilter::onDestroy, because callbacks->dispatcher().post()
    // is used to trigger callbacks from go-routines, and we have promised that
    // "Callbacks will not be invoked by the filter after onDestroy() is
    // called.". Moreover, "Every filter is responsible for making sure that any
    // async events are cleaned up in the context of this routine. This includes
    // timers, network calls, etc. [...] Filters must not invoke either encoder
    // or decoder filter callbacks after having onDestroy() invoked."
    //
    // NOTE: Already scheduled callbacks will be inhibited by remembering if
    // onDestroy() was called (e.g., "if (callbacks) cb();"). However, this
    // complicates reasoning at the scheduling site, so the decision to abort
    // is left to the callback.

    // wait for the ultimate unpin()
    Thread::LockGuard lk_m(m_);
    while (pins_.load())
      // CondVar::wait() does not throw, so it's safe to pass the mutex rather than the guard.
      cv_.wait(m_);
  }

  // release self_. This is safe now, because if there were concurrent calls to
  // post(), pins_ shouldn't be zero.
  self_.reset();

  // Fret not: we are still alive! Someone called onDestroy(), after all...

  // bluntly ensure there is no more asynchronous access
  dispatcher_ = 0;
  decoderCallbacks_ = 0;
  encoderCallbacks_ = 0;
}

void GoHttpFilter::onPost(uint64_t postTag) {
  if (!cgoTag_) {
    // TODO: Log dropped onPost()
    return;
  }

  ASSERT(cgoSafe());
  cgo_proxy_->GoHttpFilterOnPost(cgoTag_, postTag);
}

} // namespace Http
} // namespace Envoy
