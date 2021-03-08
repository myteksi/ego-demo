#pragma once
// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#ifndef CGO_ENVOY_H
#define CGO_ENVOY_H

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

#include <stddef.h>
#include <stdint.h>

// Generic C interfacing

typedef const char* GoError;

typedef struct {
  size_t len;
  char* data;
} GoStr;

typedef struct {
  size_t len;
  size_t cap;
  void* data;
} GoBuf;

// GoHttpFilter
void GoHttpFilter_pin(void* goHttpFilter);
void GoHttpFilter_unpin(void* goHttpFilter);
void GoHttpFilter_post(void* goHttpFilter, uint64_t tag);
void GoHttpFilter_log(void* goHttpFilter, uint32_t logLevel, GoStr message);
uint64_t GoHttpFilter_ResolveMostSpecificPerGoFilterConfig(void* goHttpFilter, GoStr name);

// DecoderFilterCallbacks
void GoHttpFilter_DecoderCallbacks_continueDecoding(void* goHttpFilter);
int GoHttpFilter_DecoderCallbacks_sendLocalReply(void* goHttpFilter, int responseCode, GoStr body,
                                                 GoBuf headersBuf, GoStr details);
const void* GoHttpFilter_DecoderCallbacks_decodingBuffer(void* goHttpFilter);
void GoHttpFilter_DecoderCallbacks_addDecodedData(void* goHttpFilter, void* bufferInstance,
                                                  int streamingFilter);
void GoHttpFilter_DecoderCallbacks_StreamInfo_FilterState_setData(void* goHttpFilter, GoStr name,
                                                                  GoStr value, int stateType,
                                                                  int lifeSpan);
int GoHttpFilter_DecoderCallbacks_encodeHeaders(void* goHttpFilter, int responseCode,
                                                GoBuf headersBuf, int endStream);

// EncoderFilterCallbacks
const void* GoHttpFilter_EncoderCallbacks_encodingBuffer(void* goHttpFilter);
void GoHttpFilter_EncoderCallbacks_addEncodedData(void* goHttpFilter, void* bufferInstance,
                                                  int streamingFilter);
void GoHttpFilter_EncoderCallbacks_continueEncoding(void* goHttpFilter);


// StreamFilterCallbacks
int GoHttpFilter_StreamFilterCallbacks_StreamInfo_FilterState_getDataReadOnly(void* goHttpFilter,int encoder,
                                                                         GoStr name, GoStr* value);
int64_t GoHttpFilter_StreamFilterCallbacks_StreamInfo_lastDownstreamTxByteSent(void* goHttpFilter,int encoder);
const void * GoHttpFilter_StreamFilterCallbacks_StreamInfo_getRequestHeaders(void* goHttpFilter,int encoder);     
int GoHttpFilter_StreamFilterCallbacks_StreamInfo_responseCode(void* goHttpFilter,int encoder);                                                                     
void GoHttpFilter_StreamFilterCallbacks_StreamInfo_responseCodeDetails(void* goHttpFilter,int encoder, GoStr* value);                                                                          

// Returns 0 if route isn't existing. Otherwise, returns non-zero.
int GoHttpFilter_StreamFilterCallbacks_routeExisting(void* goHttpFilter, int encoder);

// Returns 0 if there is no error. Otherwise, returns non-zero.
int GoHttpFilter_StreamFilterCallbacks_route_routeEntry_pathMatchCriterion_matcher(void* goHttpFilter, int encoder, GoStr* value);
// Returns match type as following. If there is an error, returns a negative number.
// case Envoy::Router::PathMatchType::None:
//   return 0;
//   break;
// case Envoy::Router::PathMatchType::Prefix:
//   return 1;
//   break;
// case Envoy::Router::PathMatchType::Exact:
//   return 2;
//   break;
// case Envoy::Router::PathMatchType::Regex:
//   return 3;
int GoHttpFilter_StreamFilterCallbacks_route_routeEntry_pathMatchCriterion_matchType(void* goHttpFilter, int encoder);

// GenericSecretConfigProvider
void GoHttpFilter_GenericSecretConfigProvider_secret(void* goHttpFilter, GoStr* value);


// Two specicial spanIDs. See ego/src/cc/filter/http/span-group.h
// -1 : activeSpan of decoderCallBacks
// -0: activeSpan of encoderCallbacks
// Span is used exclusively by a single Go routine.
int GoHttpFilter_Span_getContext(void *goHttpFilter, const intptr_t spanID, GoBuf buf);
intptr_t GoHttpFilter_Span_spawnChild(void *goHttpFilter, intptr_t parentSpanID,  GoStr name);
void GoHttpFilter_Span_finishSpan(void *goHttpFilter, intptr_t spanID);


uint64_t BufferInstance_copyOut(void* bufferInstance, size_t start, GoBuf buf);
uint64_t BufferInstance_length(void* bufferInstance);
uint64_t BufferInstance_getRawSlicesCount(void* bufferInstance);
uint64_t BufferInstance_getRawSlices(void* bufferInstance, uint64_t max, GoBuf* dest);

// RequestHeaderMap
void RequestHeaderMap_add(void* requestHeaderMap, GoStr name, GoStr value);
void RequestHeaderMap_set(void* requestHeaderMap, GoStr name, GoStr value);
void RequestHeaderMap_append(void* requestHeaderMap, GoStr name, GoStr value);
size_t RequestHeaderMap_getByPrefix(void* requestHeaderMap, GoStr prefix, GoBuf buf);
void RequestHeaderMap_remove(void* requestHeaderMap, GoStr name);

/**
 * Gets Path header. Header value is stored in value parameter.
 * Header value is managed by RequestHeaderMap.
 * Callers should not free memory returned in value.
 * @param requestHeaderMap pointer to Envoy::Http::RequestHeaderMap object.
 * @param value placeholder containing info about header value.
 */
void RequestHeaderMap_Path(void* requestHeaderMap, GoStr* value);

void RequestHeaderMap_setPath(void* requestHeaderMap, GoStr value);

/**
 * Gets Method header. Header value is stored in value parameter.
 * Header value is managed by RequestHeaderMap.
 * Callers should not free memory returned in value.
 * @param requestHeaderMap pointer to Envoy::Http::RequestHeaderMap object.
 * @param value placeholder containing info about header value.
 */
void RequestHeaderMap_Method(void* requestHeaderMap, GoStr* value);

/**
 * Gets ContentType header. Header value is stored in value parameter.
 * Header value is managed by RequestHeaderMap.
 * Callers should not free memory returned in value.
 * @param requestHeaderMap pointer to Envoy::Http::RequestHeaderMap object.
 * @param value placeholder containing info about header value.
 */
void RequestHeaderMap_ContentType(void* requestHeaderMap, GoStr* value);

/**
 * Gets Authorization header. Header value is stored in value parameter.
 * Header value is managed by RequestHeaderMap.
 * Callers should not free memory returned in value.
 * @param requestHeaderMap pointer to Envoy::Http::RequestHeaderMap object.
 * @param value placeholder containing info about header value.
 */
void RequestHeaderMap_Authorization(void* requestHeaderMap, GoStr* value);

/**
 * Gets header by name. Header value is stored in value parameter.
 * Header value is managed by RequestHeaderMap.
 * Callers should not free memory returned in value.
 * @param requestHeaderMap pointer to Envoy::Http::RequestHeaderMap object.
 * @param name header name. It's caller's responsibility to manage name's memory.
 * @param value placeholder containing info about header value.
 */
void RequestHeaderMap_get(void* requestHeaderMap, GoStr name, GoStr* value);

// RequestTrailerMap
void RequestTrailerMap_add(void* requestTrailerMap, GoStr name, GoStr value);

// ResponseHeaderMap
void ResponseHeaderMap_add(void* responseHeaderMap, GoStr name, GoStr value);
void ResponseHeaderMap_set(void* responseHeaderMap, GoStr name, GoStr value);
void ResponseHeaderMap_append(void* responseHeaderMap, GoStr name, GoStr value);
void ResponseHeaderMap_remove(void* responseHeaderMap, GoStr name);
void ResponseHeaderMap_get(void* responseHeaderMap, GoStr name, GoStr* value);
void ResponseHeaderMap_ContentType(void* responseHeaderMap, GoStr* value);
void ResponseHeaderMap_Status(void* responseHeaderMap, GoStr* value);
void ResponseHeaderMap_setStatus(void* responseHeaderMap, int status);

// Static functions will be call from from Go ("downcalls") without a pointer
//
void Envoy_log_misc(uint32_t level, GoStr tag, GoStr message);

// Stats::Scope
const void* Stats_Scope_counterFromStatName(void* scope, GoStr name);
const void* Stats_Scope_gaugeFromStatName(void* scope, GoStr name, int importMode);
const void* Stats_Scope_histogramFromStatName(void* scope, GoStr name, int unit);

// Stats::Counter
void Stats_Counter_add(void* counter, uint64_t amount);
void Stats_Counter_inc(void* counter);
uint64_t Stats_Counter_latch(void* counter);
void Stats_Counter_reset(void* counter);
uint64_t Stats_Counter_value(void* counter);

// Stats::Gauge
void Stats_Gauge_add(void* gauge, uint64_t amount);
void Stats_Gauge_dec(void* gauge);
void Stats_Gauge_inc(void* gauge);
void Stats_Gauge_set(void* gauge, uint64_t value);
void Stats_Gauge_sub(void* gauge, uint64_t amount);
uint64_t Stats_Gauge_value(void* gauge);

// Stats::Histogram
int Stats_Histogram_unit(void* histogram);
void Stats_Histogram_recordValue(void* histogram, uint64_t value);

#ifdef __cplusplus
}
#endif // __cplusplus
#endif // CGO_ENVOY_H
