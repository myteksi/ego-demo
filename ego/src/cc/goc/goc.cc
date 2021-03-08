// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "goc.h"

Envoy::Http::FilterHeadersStatus Goc_FilterHeadersStatus(int status) {
  switch (status) {
  case 100:
    return Envoy::Http::FilterHeadersStatus::Continue;
  case 101:
    return Envoy::Http::FilterHeadersStatus::StopIteration;
  case 102:
    return Envoy::Http::FilterHeadersStatus::ContinueAndEndStream;
  case 104:
    return Envoy::Http::FilterHeadersStatus::StopAllIterationAndBuffer;
  case 105:
    return Envoy::Http::FilterHeadersStatus::StopAllIterationAndWatermark;
  default:
    // TODO: log error
    return Envoy::Http::FilterHeadersStatus::StopIteration;
  }
}

Envoy::Http::FilterTrailersStatus Goc_FilterTrailersStatus(int status) {
  switch (status) {
  case 200:
    return Envoy::Http::FilterTrailersStatus::Continue;
  case 201:
    return Envoy::Http::FilterTrailersStatus::StopIteration;
  default:
    // TODO: log error
    return Envoy::Http::FilterTrailersStatus::StopIteration;
  }
}

Envoy::Http::FilterDataStatus Goc_FilterDataStatus(int status) {
  switch (status) {
  case 300:
    return Envoy::Http::FilterDataStatus::Continue;
  case 301:
    return Envoy::Http::FilterDataStatus::StopIterationAndBuffer;
  case 302:
    return Envoy::Http::FilterDataStatus::StopIterationAndWatermark;
  case 303:
    return Envoy::Http::FilterDataStatus::StopIterationNoBuffer;
  default:
    // TODO: log error
    return Envoy::Http::FilterDataStatus::StopIterationNoBuffer;
  }
}

Envoy::Http::Code Goc_HttpResonseCode(int status) {
  switch (status) {
  case 100:
    return Envoy::Http::Code::Continue;
  case 101:
    return Envoy::Http::Code::SwitchingProtocols;

  case 200:
    return Envoy::Http::Code::OK;
  case 201:
    return Envoy::Http::Code::Created;
  case 202:
    return Envoy::Http::Code::Accepted;
  case 203:
    return Envoy::Http::Code::NonAuthoritativeInformation;
  case 204:
    return Envoy::Http::Code::NoContent;
  case 205:
    return Envoy::Http::Code::ResetContent;
  case 206:
    return Envoy::Http::Code::PartialContent;
  case 207:
    return Envoy::Http::Code::MultiStatus;
  case 208:
    return Envoy::Http::Code::AlreadyReported;
  case 226:
    return Envoy::Http::Code::IMUsed;

  case 300:
    return Envoy::Http::Code::MultipleChoices;
  case 301:
    return Envoy::Http::Code::MovedPermanently;
  case 302:
    return Envoy::Http::Code::Found;
  case 303:
    return Envoy::Http::Code::SeeOther;
  case 304:
    return Envoy::Http::Code::NotModified;
  case 305:
    return Envoy::Http::Code::UseProxy;
  case 306:
    return Envoy::Http::Code::TemporaryRedirect;
  case 307:
    return Envoy::Http::Code::PermanentRedirect;

  case 400:
    return Envoy::Http::Code::BadRequest;
  case 401:
    return Envoy::Http::Code::Unauthorized;
  case 402:
    return Envoy::Http::Code::PaymentRequired;
  case 403:
    return Envoy::Http::Code::Forbidden;
  case 404:
    return Envoy::Http::Code::NotFound;
  case 405:
    return Envoy::Http::Code::MethodNotAllowed;
  case 406:
    return Envoy::Http::Code::NotAcceptable;
  case 407:
    return Envoy::Http::Code::ProxyAuthenticationRequired;
  case 408:
    return Envoy::Http::Code::RequestTimeout;
  case 409:
    return Envoy::Http::Code::Conflict;
  case 410:
    return Envoy::Http::Code::Gone;
  case 411:
    return Envoy::Http::Code::LengthRequired;
  case 412:
    return Envoy::Http::Code::PreconditionFailed;
  case 413:
    return Envoy::Http::Code::PayloadTooLarge;
  case 414:
    return Envoy::Http::Code::URITooLong;
  case 415:
    return Envoy::Http::Code::UnsupportedMediaType;
  case 416:
    return Envoy::Http::Code::RangeNotSatisfiable;
  case 417:
    return Envoy::Http::Code::ExpectationFailed;
  case 421:
    return Envoy::Http::Code::MisdirectedRequest;
  case 422:
    return Envoy::Http::Code::UnprocessableEntity;
  case 423:
    return Envoy::Http::Code::Locked;
  case 424:
    return Envoy::Http::Code::FailedDependency;
  case 426:
    return Envoy::Http::Code::UpgradeRequired;
  case 428:
    return Envoy::Http::Code::PreconditionRequired;
  case 429:
    return Envoy::Http::Code::TooManyRequests;
  case 431:
    return Envoy::Http::Code::RequestHeaderFieldsTooLarge;

  case 500:
    return Envoy::Http::Code::InternalServerError;
  case 501:
    return Envoy::Http::Code::NotImplemented;
  case 502:
    return Envoy::Http::Code::BadGateway;
  case 503:
    return Envoy::Http::Code::ServiceUnavailable;
  case 504:
    return Envoy::Http::Code::GatewayTimeout;
  case 505:
    return Envoy::Http::Code::HTTPVersionNotSupported;
  case 506:
    return Envoy::Http::Code::VariantAlsoNegotiates;
  case 507:
    return Envoy::Http::Code::InsufficientStorage;
  case 508:
    return Envoy::Http::Code::LoopDetected;
  case 510:
    return Envoy::Http::Code::NotExtended;
  case 511:
    return Envoy::Http::Code::NetworkAuthenticationRequired;

  default:
    // TODO: log error
    return Envoy::Http::Code::InternalServerError;
  }
}

Envoy::StreamInfo::FilterState::StateType Goc_FilterStateType(int stateType) {
  switch (stateType) {
  case 1:
    return Envoy::StreamInfo::FilterState::StateType::ReadOnly;
  case 2:
    return Envoy::StreamInfo::FilterState::StateType::Mutable;

  default:
    // TODO: log error
    return Envoy::StreamInfo::FilterState::StateType::ReadOnly;
  }
}

Envoy::StreamInfo::FilterState::LifeSpan Goc_FilterStateLifeSpan(int lifeSpan) {
  switch (lifeSpan) {
  case 1:
    return Envoy::StreamInfo::FilterState::LifeSpan::FilterChain;
  case 2:
    return Envoy::StreamInfo::FilterState::LifeSpan::DownstreamRequest;
  case 3:
    return Envoy::StreamInfo::FilterState::LifeSpan::DownstreamConnection;
  case 4:
    return Envoy::StreamInfo::FilterState::LifeSpan::TopSpan;

  default:
    // TODO: log error
    return Envoy::StreamInfo::FilterState::LifeSpan::FilterChain;
  }
}

Envoy::Stats::Gauge::ImportMode Goc_Stats_ImportMode(int importMode) {
  switch (importMode) {
  case 1:
    return Envoy::Stats::Gauge::ImportMode::Uninitialized;
  case 2:
    return Envoy::Stats::Gauge::ImportMode::NeverImport;
  case 3:
    return Envoy::Stats::Gauge::ImportMode::Accumulate;
  default:
    // TODO: log error
    return Envoy::Stats::Gauge::ImportMode::Uninitialized;
  }
}

Envoy::Stats::Histogram::Unit Goc_Stats_Unit(int unit) {
  switch (unit) {
  case 1:
    return Envoy::Stats::Histogram::Unit::Null;
  case 2:
    return Envoy::Stats::Histogram::Unit::Unspecified;
  case 3:
    return Envoy::Stats::Histogram::Unit::Bytes;
  case 4:
    return Envoy::Stats::Histogram::Unit::Microseconds;
  case 5:
    return Envoy::Stats::Histogram::Unit::Milliseconds;
  default:
    // TODO: log error
    return Envoy::Stats::Histogram::Unit::Null;
  }
}

int Goc_Stats_Unit_Value(Envoy::Stats::Histogram::Unit unit) {
  switch (unit) {
  case Envoy::Stats::Histogram::Unit::Null:
    return 1;
  case Envoy::Stats::Histogram::Unit::Unspecified:
    return 2;
  case Envoy::Stats::Histogram::Unit::Bytes:
    return 3;
  case Envoy::Stats::Histogram::Unit::Microseconds:
    return 4;
  case Envoy::Stats::Histogram::Unit::Milliseconds:
    return 5;
  default:
    // TODO: log error
    return 0;
  }
}
