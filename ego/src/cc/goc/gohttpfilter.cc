// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "common/common/empty_string.h"
#include "common/config/datasource.h"
#include "common/http/header_map_impl.h"
#include "common/router/string_accessor_impl.h"

#include "ego/src/cc/filter/http/filter.h"
#include "ego/src/cc/goc/proto/dto.pb.validate.h"
#include "envoy.h"
#include "goc.h"

void GoHttpFilter_pin(void* goHttpFilter) {
  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->pin();
}

void GoHttpFilter_unpin(void* goHttpFilter) {
  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->unpin();
}

void GoHttpFilter_post(void* goHttpFilter, uint64_t tag) {
  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->post(tag);
}

void GoHttpFilter_log(void* goHttpFilter, uint32_t logLevel, GoStr messsage) {

  auto c_message = absl::string_view(messsage.data, messsage.len);

  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->log(logLevel, c_message);
}

void GoHttpFilter_DecoderCallbacks_continueDecoding(void* goHttpFilter) {
  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->decoderCallbacks()->continueDecoding();
}

int GoHttpFilter_DecoderCallbacks_sendLocalReply(void* goHttpFilter, int responseCode, GoStr body,
                                                 GoBuf headersBuf, GoStr details) {
  auto c_body = absl::string_view(body.data, body.len);
  auto c_details = absl::string_view(details.data, details.len);
  auto headers = ego::http::RequestHeaderMap{};
  if (!headers.ParseFromArray(headersBuf.data, headersBuf.len)) {
    // non-zero returned value means errors.
    return 1;
  }

  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
      ->decoderCallbacks()
      ->sendLocalReply(
          Goc_HttpResonseCode(responseCode), c_body,
          [headers](Envoy::Http::HeaderMap& response_headers) -> void {
            for (const auto& h : headers.headers()) {
              response_headers.setCopy(Envoy::Http::LowerCaseString(h.key()), h.value());
            }
          },
          absl::nullopt /*grpc_status*/, c_details);
  return 0;
}

void GoHttpFilter_DecoderCallbacks_StreamInfo_FilterState_setData(void* goHttpFilter, GoStr name,
                                                                  GoStr value, int stateType,
                                                                  int lifeSpan) {
  auto c_name = absl::string_view(name.data, name.len);
  auto c_value = absl::string_view(value.data, value.len);

  // StringAccessorImpl copies c_value data
  std::shared_ptr<Envoy::StreamInfo::FilterState::Object> value_object =
      std::make_unique<Envoy::Router::StringAccessorImpl>(c_value);

  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
      ->decoderCallbacks()
      ->streamInfo()
      .filterState()
      ->setData(c_name, value_object, Goc_FilterStateType(stateType),
                Goc_FilterStateLifeSpan(lifeSpan));
}

int GoHttpFilter_StreamFilterCallbacks_StreamInfo_FilterState_getDataReadOnly(void* goHttpFilter, int encoder,
                                                                         GoStr name, GoStr* value) {
  ASSERT(nullptr != goHttpFilter);
  ASSERT(nullptr != value);

  auto c_name = absl::string_view(name.data, name.len);

  auto filter_state = static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
                          ->streamFilterCallbacks(encoder)
                          ->streamInfo()
                          .filterState();
  if (!filter_state->hasData<Envoy::Router::StringAccessorImpl>(c_name)) {
    return 0;
  }

  auto c_value =
      filter_state->getDataReadOnly<Envoy::Router::StringAccessorImpl>(c_name).asString();
  value->len = c_value.size();
  value->data = const_cast<char*>(c_value.data());
  return 1;
}

void GoHttpFilter_StreamFilterCallbacks_StreamInfo_responseCodeDetails(void* goHttpFilter, int encoder, GoStr* value) {
  ASSERT(nullptr != goHttpFilter);
  ASSERT(nullptr != value);

  auto & response_code_details = static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
                                   ->streamFilterCallbacks(encoder)
                                   ->streamInfo()
                                   .responseCodeDetails()
                                   .value();
  value->len = response_code_details.size();
  value->data = const_cast<char*>(response_code_details.data());
}

int64_t GoHttpFilter_StreamFilterCallbacks_StreamInfo_lastDownstreamTxByteSent(void* goHttpFilter,int encoder){
  auto request_time = static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
                        ->streamFilterCallbacks(encoder)
                        ->streamInfo()
                        .lastDownstreamTxByteSent();
  return request_time.value_or(std::chrono::nanoseconds(-1)).count();
}

const void * GoHttpFilter_StreamFilterCallbacks_StreamInfo_getRequestHeaders(void* goHttpFilter,int encoder){
  return static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
                        ->streamFilterCallbacks(encoder)
                        ->streamInfo()
                        .getRequestHeaders();
}

int GoHttpFilter_StreamFilterCallbacks_StreamInfo_responseCode(void* goHttpFilter,int encoder){
  return static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
                        ->streamFilterCallbacks(encoder)
                        ->streamInfo()
                        .responseCode().value_or(0);
}

int GoHttpFilter_StreamFilterCallbacks_routeExisting(void* goHttpFilter, int encoder) {
  ASSERT(nullptr != goHttpFilter);

  auto route =
      static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->streamFilterCallbacks(encoder)->route();
  return nullptr == route? 0: 1;
}

int GoHttpFilter_StreamFilterCallbacks_route_routeEntry_pathMatchCriterion_matcher(void* goHttpFilter, int encoder, GoStr* value) {
  ASSERT(nullptr != goHttpFilter);
  ASSERT(nullptr != value);

  auto route =
      static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->streamFilterCallbacks(encoder)->route();
  if (nullptr == route || nullptr == route->routeEntry()) {
    return 1;
  }

  const auto &matcher = route->routeEntry()->pathMatchCriterion().matcher();
  value->len = matcher.size();
  value->data = const_cast<char*>(matcher.data());
  return 0;
}

int GoHttpFilter_StreamFilterCallbacks_route_routeEntry_pathMatchCriterion_matchType(void* goHttpFilter, int encoder) {
  ASSERT(nullptr != goHttpFilter);

  auto route =
      static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->streamFilterCallbacks(encoder)->route();
  if (nullptr == route || nullptr == route->routeEntry()) {
    return -1;
  }

  auto matchType = route->routeEntry()->pathMatchCriterion().matchType();
  switch(matchType) {
    case Envoy::Router::PathMatchType::None:
      return 0;
      break;
    case Envoy::Router::PathMatchType::Prefix:
      return 1;
      break;
    case Envoy::Router::PathMatchType::Exact:
      return 2;
      break;
    case Envoy::Router::PathMatchType::Regex:
      return 3;
      break;
  }
  return -1;
}

const void* GoHttpFilter_DecoderCallbacks_decodingBuffer(void* goHttpFilter) {
  return static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
      ->decoderCallbacks()
      ->decodingBuffer();
}

void GoHttpFilter_DecoderCallbacks_addDecodedData(void* goHttpFilter, void* bufferInstance,
                                                  int streamingFilter) {

  auto c_bufferInstance = static_cast<Envoy::Buffer::Instance*>(bufferInstance);

  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
      ->decoderCallbacks()
      ->addDecodedData(*c_bufferInstance, streamingFilter);
}

int GoHttpFilter_DecoderCallbacks_encodeHeaders(void* goHttpFilter, int responseCode,
                                                GoBuf headersBuf, int endStream) {
  auto headers = ego::http::ResponseHeaderMap{};
  if (!headers.ParseFromArray(headersBuf.data, headersBuf.len)) {
    // non-zero returned value means errors.
    return 1;
  }

  auto response_headers{Envoy::Http::createHeaderMap<Envoy::Http::ResponseHeaderMapImpl>(
      {{Envoy::Http::Headers::get().Status, std::to_string(responseCode)}})};

  for (const auto& h : headers.headers()) {
    response_headers->addCopy(Envoy::Http::LowerCaseString(h.key()), h.value());
  }

  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
      ->decoderCallbacks()
      ->encodeHeaders(std::move(response_headers), endStream == 1 ? true : false);
  return 0;
}

const void* GoHttpFilter_EncoderCallbacks_encodingBuffer(void* goHttpFilter) {
  return static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
      ->encoderCallbacks()
      ->encodingBuffer();
}

void GoHttpFilter_EncoderCallbacks_addEncodedData(void* goHttpFilter, void* bufferInstance,
                                                  int streamingFilter) {

  auto c_bufferInstance = static_cast<Envoy::Buffer::Instance*>(bufferInstance);

  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
      ->encoderCallbacks()
      ->addEncodedData(*c_bufferInstance, streamingFilter);
}

void GoHttpFilter_EncoderCallbacks_continueEncoding(void* goHttpFilter) {
  static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)->encoderCallbacks()->continueEncoding();
}

void GoHttpFilter_GenericSecretConfigProvider_secret(void* goHttpFilter, GoStr* value) {
  auto filter = static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter);
  auto secretProvider = filter->genericSecretConfigProvider();
  if (secretProvider == nullptr || secretProvider->secret() == nullptr) {
    return;
  }
  // We need a variable on C-Side to hold the reference to not free after return
  filter->secret_holder =
      Envoy::Config::DataSource::read(secretProvider->secret()->secret(), true, filter->api());

  value->len = filter->secret_holder.size();
  value->data = const_cast<char*>(filter->secret_holder.c_str());
}

int GoHttpFilter_Span_getContext(void *goHttpFilter, const intptr_t spanID, GoBuf buf){
    auto that = static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter);

    Envoy::Http::RequestHeaderMapImpl headers;
    that->getSpan(spanID).injectContext(headers);
    auto result = ego::http::RequestHeaderMap{};

    headers.iterate([](const Envoy::Http::HeaderEntry& header, void* context) -> Envoy::Http::HeaderMap::Iterate {
      auto r = static_cast<ego::http::RequestHeaderMap*>(context);
      auto entry = r->mutable_headers()->Add();
      entry->set_key(header.key().getStringView().data(), header.key().getStringView().size());
      entry->set_value(header.value().getStringView().data(), header.value().getStringView().size());

      return Envoy::Http::HeaderMap::Iterate::Continue;
    }, &result);

    // no headers matches the prefix
    if(0 == result.headers_size()) {
        return 0;
    }

    // the buffer is too small. Return required size.
    const auto size = result.ByteSizeLong();
    if (size > buf.len) {
        return size;
    }

    // serialize data.
    result.SerializePartialToArray(buf.data, buf.len);
    return size;
}

intptr_t GoHttpFilter_Span_spawnChild(void *goHttpFilter, intptr_t parentSpanID, GoStr name) {
  auto that = static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter);
  auto c_name = std::string(name.data, name.len);

  return that->spawnChildSpan(parentSpanID, c_name);
}

void GoHttpFilter_Span_finishSpan(void *goHttpFilter, intptr_t spanID) {
  auto that = static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter);

  that->finishSpan(spanID);
}


// Not use name GoStr for now, assumming that Go filters always get their own configurations.
uint64_t GoHttpFilter_ResolveMostSpecificPerGoFilterConfig(void* goHttpFilter, GoStr) {
  return static_cast<Envoy::Http::GoHttpFilter*>(goHttpFilter)
      ->resolveMostSpecificPerGoFilterConfigTag();
}