#pragma once
// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include <memory>

namespace Envoy {
namespace Http {

class CgoProxy {
public:
  virtual ~CgoProxy();

  virtual unsigned long long GoHttpFilterCreate(void* native, unsigned long long factory_tag,
                                                unsigned long long filter_slot) = 0;
  virtual void GoHttpFilterOnDestroy(unsigned long long filter_tag) = 0;
  virtual long long GoHttpFilterDecodeHeaders(unsigned long long filter_tag, void* headers,
                                              int end_stream) = 0;
  virtual long long GoHttpFilterDecodeData(unsigned long long filter_tag, void* buffer,
                                           int end_stream) = 0;
  virtual long long GoHttpFilterDecodeTrailers(unsigned long long filter_tag, void* trailers) = 0;
  virtual long long GoHttpFilterEncodeHeaders(unsigned long long filter_tag, void* headers,
                                              int end_stream) = 0;
  virtual long long GoHttpFilterEncodeData(unsigned long long filter_tag, void* headers,
                                           int end_stream) = 0;
  virtual void GoHttpFilterOnPost(unsigned long long filter_tag, unsigned long long post_tag) = 0;
};

class CgoProxyImpl : public CgoProxy {
public:
  CgoProxyImpl();
  ~CgoProxyImpl() override;

  unsigned long long GoHttpFilterCreate(void* native, unsigned long long factory_tag,
                                        unsigned long long filter_slot) override;
  void GoHttpFilterOnDestroy(unsigned long long filter_tag) override;
  long long GoHttpFilterDecodeHeaders(unsigned long long filter_tag, void* headers,
                                      int end_stream) override;
  long long GoHttpFilterDecodeData(unsigned long long filter_tag, void* buffer,
                                   int end_stream) override;
  long long GoHttpFilterDecodeTrailers(unsigned long long filter_tag, void* trailers) override;
  long long GoHttpFilterEncodeHeaders(unsigned long long filter_tag, void* headers,
                                      int end_stream) override;
  long long GoHttpFilterEncodeData(unsigned long long filter_tag, void* headers,
                                   int end_stream) override;
  void GoHttpFilterOnPost(unsigned long long filter_tag, unsigned long long post_tag) override;
};

using CgoProxyPtr = std::shared_ptr<CgoProxy>;
} // namespace Http
} // namespace Envoy