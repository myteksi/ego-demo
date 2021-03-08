#pragma once
// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#ifndef FILTER_HTTP_GETHEADER_FILTER_H
#define FILTER_HTTP_GETHEADER_FILTER_H

#include <atomic>

#include "envoy/server/filter_config.h"
#include "envoy/stats/scope.h"

#include "common/common/thread.h"
#include "common/singleton/const_singleton.h"

#include "cgo-proxy.h"
#include "ego/src/cc/filter/http/filter.pb.h"
#include "span-group.h"

namespace Envoy {
namespace Http {

struct GoHttpConstantValues {
  const std::string FilterName = "ego_http";
};

using GoHttpConstants = ConstSingleton<GoHttpConstantValues>;

// This class represents the proto configuration declared in filter.proto
//
class GoHttpFilterConfig : public Logger::Loggable<Logger::Id::config> {
public:
  GoHttpFilterConfig(const ego::http::Settings& proto, const Stats::ScopeSharedPtr& scope);
  ~GoHttpFilterConfig();

private:
  friend class GoHttpFilter;
  uint64_t cgoTag_;

  uint64_t cgoSlot_;
  inline bool cgoSafe();

  const std::string filter_;
  std::string filter() const { return filter_; }

  // hold the scope_ for using from go side
  Stats::ScopeSharedPtr scope_;
};

class GoHttpRouteSpecificFilterConfig : public Router::RouteSpecificFilterConfig {
public:
  GoHttpRouteSpecificFilterConfig(const ego::http::SettingsPerRoute& config);
  ~GoHttpRouteSpecificFilterConfig();

  uint64_t cgoTag(std::string name) const;

private:
  uint64_t cgoSlot_;
  inline bool cgoSafe();

  std::map<std::string, uint64_t> filters_;
  std::function<void()> onDestroy_;
};

// This class implements the actual filter logic
//
class GoHttpFilter : public StreamFilter, public Logger::Loggable<Logger::Id::filter> {
public:
  GoHttpFilter(std::shared_ptr<GoHttpFilterConfig> config, Api::Api& api,
               Secret::GenericSecretConfigProviderSharedPtr secret_provider, CgoProxyPtr cgo_proxy, SpanGroupPtr span_group);
  ~GoHttpFilter() override{};

  Http::StreamFilterSharedPtr ref() { return self_; }
  void pin();
  void unpin();
  void post(uint64_t tag);
  void log(uint32_t level, absl::string_view message);

  // Public decoderCallbacks_ to let GOC calling to continueDecoding, sendLocalReply, ...
  StreamDecoderFilterCallbacks* decoderCallbacks();

  // Public encoderCallbacks_ to let GOC calling to continueEncoding, ...
  StreamEncoderFilterCallbacks* encoderCallbacks();

  StreamFilterCallbacks* streamFilterCallbacks(int encoder);

  // Public interface to access secret provider from Go-side
  Secret::GenericSecretConfigProviderSharedPtr genericSecretConfigProvider();
  Api::Api& api();
  std::string secret_holder;

public:
  // Http::StreamFilterBase
  void onDestroy() override;

  // Http::StreamDecoderFilter
  FilterHeadersStatus decodeHeaders(RequestHeaderMap&, bool) override;
  FilterDataStatus decodeData(Buffer::Instance&, bool) override;
  FilterTrailersStatus decodeTrailers(RequestTrailerMap&) override;
  void setDecoderFilterCallbacks(StreamDecoderFilterCallbacks&) override;

  // Http::StreamEncoderFilter

  /**
   * Called with 100-continue headers.
   *
   * This is not folded into encodeHeaders because most Envoy users and filters
   * will not be proxying 100-continue and with it split out, can ignore the
   * complexity of multiple encodeHeaders calls.
   *
   * @param headers supplies the 100-continue response headers to be encoded.
   * @return FilterHeadersStatus determines how filter chain iteration proceeds.
   *
   */
  FilterHeadersStatus encode100ContinueHeaders(ResponseHeaderMap& headers) override;

  /**
   * Called with headers to be encoded, optionally indicating end of stream.
   * @param headers supplies the headers to be encoded.
   * @param end_stream supplies whether this is a header only request/response.
   * @return FilterHeadersStatus determines how filter chain iteration proceeds.
   */
  FilterHeadersStatus encodeHeaders(ResponseHeaderMap& headers, bool end_stream) override;

  /**
   * Called with data to be encoded, optionally indicating end of stream.
   * @param data supplies the data to be encoded.
   * @param end_stream supplies whether this is the last data frame.
   * @return FilterDataStatus determines how filter chain iteration proceeds.
   */
  FilterDataStatus encodeData(Buffer::Instance& data, bool end_stream) override;

  /**
   * Called with trailers to be encoded, implicitly ending the stream.
   * @param trailers supplies the trailers to be encoded.
   */
  FilterTrailersStatus encodeTrailers(ResponseTrailerMap& trailers) override;

  /**
   * Called with metadata to be encoded. New metadata should be added directly to metadata_map. DO
   * NOT call StreamDecoderFilterCallbacks::encodeMetadata() interface to add new metadata.
   *
   * @param metadata_map supplies the metadata to be encoded.
   * @return FilterMetadataStatus, which currently is always FilterMetadataStatus::Continue;
   */
  FilterMetadataStatus encodeMetadata(MetadataMap& metadata_map) override;

  /**
   * Called by the filter manager once to initialize the filter callbacks that the filter should
   * use. Callbacks will not be invoked by the filter after onDestroy() is called.
   */
  void setEncoderFilterCallbacks(StreamEncoderFilterCallbacks& callbacks) override;

  /**
   * Called at the end of the stream, when all data has been encoded.
   */
  void encodeComplete() override;

  // gets route specific filter config cgo tag.
  uint64_t resolveMostSpecificPerGoFilterConfigTag();

  intptr_t spawnChildSpan(const intptr_t parent_span_id, std::string &name);
  void finishSpan(const intptr_t span_id);
  Envoy::Tracing::Span& getSpan(const intptr_t span_id);


private:
  // config containing the few bits interesting on the C++ side of things
  const std::shared_ptr<GoHttpFilterConfig> config_;

  // Do only access from dispatcher context. Do check if non-0 before use.
  StreamDecoderFilterCallbacks* decoderCallbacks_;

  // Do only access from dispatcher context. Do check if non-0 before use.
  StreamEncoderFilterCallbacks* encoderCallbacks_;

  // Do only access from dispatcher context and for calling post().
  // Do check if non-0 before use.
  Event::Dispatcher* dispatcher_;

  // the ID of the Go filter object kept alive by the clutch kludge.
  uint64_t cgoTag_;

  // ref counting state for asynchronous requests. We trust this is upped
  // before every go-routine start and decreased every time a filter go
  // routine returns.
  std::atomic<int> pins_;

  // This one is only needed for keeping the filter object alive in case of
  // scheduled post() callbacks. The ref counter is shared with the filter
  // factory's call to addStreamDecoderFilter().
  Http::StreamFilterSharedPtr self_;

  // C++11 semaphore surrogate
  Thread::MutexBasicLockable m_;
  Thread::CondVar cv_;
  void done();

  // onPost is virtual to work around dependency cycles: the implementation
  // performs an upcall to a Go function, but it also needs to be referenced
  // by post() when it schedules the callback. post() is invoked via a
  // downcall from Go, and thus, the Go library becomes a cyclic dependency.
  // Making onPost virtual relaxes this because now, post only needs to know
  // the location of onPost in the virtual function table, which can be
  // determined based on the header file only.
  virtual void onPost(uint64_t tag);

  // we're still learning, so better check twice
  uint64_t cgoSlot_;
  inline bool cgoSafe();

  // hold the secret provider on C-side
  // and provide a interface to access secret from Go-side
  Api::Api& api_;
  Secret::GenericSecretConfigProviderSharedPtr secret_provider_;

  // private x-request-id for logging
  absl::string_view x_request_id_ = "";

  CgoProxyPtr cgo_proxy_;

  SpanGroupPtr span_group_;
};

} // namespace Http
} // namespace Envoy

#endif // FILTER_HTTP_GETHEADER_FILTER_H
