// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "envoy/http/header_map.h"

#include "envoy.h"

void RequestTrailerMap_add(void* requestTrailerMap, GoStr name, GoStr value) {

  auto that = static_cast<Envoy::Http::RequestTrailerMap*>(requestTrailerMap);

  // The std::string() constructor will create a copy of name. Unfortunately,
  // there is no std::string_view, because...
  auto c_name = std::string(name.data, name.len);

  // ...LowerCaseString creates a wrapped lowercase copy of c_name.
  auto w_name = Envoy::Http::LowerCaseString(c_name);

  // we are wrapping the header value with a lightweight string_view, which
  // is safe because...
  auto w_value = absl::string_view(value.data, value.len);

  // ...addCopy will add another header `w_name` associated with a copy of
  // `w_value`
  that->addCopy(w_name, w_value);
}