// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "gmock/gmock.h" // Brings in gMock.

namespace Envoy {
namespace Http {

class MockCgoProxy : public CgoProxy {
public:
  MockCgoProxy() = default;
  ~MockCgoProxy() override = default;

  MOCK_METHOD(unsigned long long, GoHttpFilterCreate,
              (void* native, unsigned long long factory_tag, unsigned long long filterSlot),
              (override));
  MOCK_METHOD(void, GoHttpFilterOnDestroy, (unsigned long long filter_tag), (override));
  MOCK_METHOD(long long, GoHttpFilterDecodeHeaders,
              (unsigned long long filter_tag, void* headers, int end_stream), (override));
  MOCK_METHOD(long long, GoHttpFilterDecodeData,
              (unsigned long long filter_tag, void* buffer, int end_stream), (override));
  MOCK_METHOD(long long, GoHttpFilterDecodeTrailers,
              (unsigned long long filter_tag, void* trailers), (override));
  MOCK_METHOD(long long, GoHttpFilterEncodeHeaders,
              (unsigned long long filter_tag, void* headers, int end_stream), (override));
  MOCK_METHOD(long long, GoHttpFilterEncodeData,
              (unsigned long long filter_tag, void* headers, int end_stream), (override));
  MOCK_METHOD(void, GoHttpFilterOnPost,
              (unsigned long long filter_tag, unsigned long long post_tag), (override));
};

} // namespace Http
} // namespace Envoy