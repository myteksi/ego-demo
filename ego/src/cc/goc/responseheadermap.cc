// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "envoy/http/header_map.h"

#include "absl/strings/match.h"
#include "ego/src/cc/goc/proto/dto.pb.validate.h"
#include "envoy.h"

void ResponseHeaderMap_add(void* responseHeaderMap, GoStr name, GoStr value) {
  auto that = static_cast<Envoy::Http::ResponseHeaderMap*>(responseHeaderMap);

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

void ResponseHeaderMap_set(void* responseHeaderMap, GoStr name, GoStr value) {
  auto that = static_cast<Envoy::Http::ResponseHeaderMap*>(responseHeaderMap);

  // The std::string() constructor will create a copy of name. Unfortunately,
  // there is no std::string_view, because...
  auto c_name = std::string(name.data, name.len);

  // ...LowerCaseString creates a wrapped lowercase copy of c_name.
  auto w_name = Envoy::Http::LowerCaseString(c_name);

  // we are wrapping the header value with a lightweight string_view, which
  // is safe because...
  auto w_value = absl::string_view(value.data, value.len);

  // ...setCopy will add another header `w_name` associated with a copy of
  // `w_value`
  that->setCopy(w_name, w_value);
}

void ResponseHeaderMap_append(void* responseHeaderMap, GoStr name, GoStr value) {
  auto that = static_cast<Envoy::Http::ResponseHeaderMap*>(responseHeaderMap);

  // The std::string() constructor will create a copy of name. Unfortunately,
  // there is no std::string_view, because...
  auto c_name = std::string(name.data, name.len);

  // ...LowerCaseString creates a wrapped lowercase copy of c_name.
  auto w_name = Envoy::Http::LowerCaseString(c_name);

  // we are wrapping the header value with a lightweight string_view, which
  // is safe because...
  auto w_value = absl::string_view(value.data, value.len);

  // ...appendCopy will add another header `w_name` associated with a copy of
  // `w_value`
  that->appendCopy(w_name, w_value);
}

void ResponseHeaderMap_remove(void* responseHeaderMap, GoStr name) {
  auto that = static_cast<Envoy::Http::ResponseHeaderMap*>(responseHeaderMap);

  // The std::string() constructor will create a copy of name. Unfortunately,
  // there is no std::string_view, because...
  auto c_name = std::string(name.data, name.len);

  // ...LowerCaseString creates a wrapped lowercase copy of c_name.
  auto w_name = Envoy::Http::LowerCaseString(c_name);
  that->remove(w_name);
}

void ResponseHeaderMap_ContentType(void* responseHeaderMap, GoStr* value) {
  ASSERT(nullptr != responseHeaderMap);
  ASSERT(nullptr != value);

  auto that = static_cast<Envoy::Http::ResponseHeaderMap*>(responseHeaderMap);

  if (nullptr == that->ContentType()) {
    return;
  }
  auto contentType = that->ContentType()->value().getStringView();

  // ContentType() returns a pointer, value() returns a reference,
  // therefore getStringView().data() should be valid after return
  value->len = contentType.size();
  value->data = const_cast<char*>(contentType.data());
}

void ResponseHeaderMap_get(void* responseHeaderMap, GoStr key, GoStr* value) {
  ASSERT(nullptr != responseHeaderMap);
  ASSERT(nullptr != value);

  auto that = static_cast<Envoy::Http::ResponseHeaderMap*>(responseHeaderMap);

  auto c_name = std::string(key.data, key.len);
  auto w_name = Envoy::Http::LowerCaseString(c_name);

  if (that->get(w_name) == nullptr) {
    return;
  }

  auto valStringView = that->get(w_name)->value().getStringView();

  // get() returns a pointer, value() returns a reference,
  // therefore getStringView().data() should be valid after return
  value->len = valStringView.size();
  value->data = const_cast<char*>(valStringView.data());
}

void ResponseHeaderMap_Status(void* responseHeaderMap, GoStr* value) {
  ASSERT(nullptr != responseHeaderMap);
  ASSERT(nullptr != value);

  auto that = static_cast<Envoy::Http::ResponseHeaderMap*>(responseHeaderMap);

  ASSERT(nullptr != that->Status());
  auto status = that->Status()->value().getStringView();

  // Status() returns a pointer, value() returns a reference,
  // therefore getStringView().data() should be valid after return
  value->len = status.size();
  value->data = const_cast<char*>(status.data());
}

void ResponseHeaderMap_setStatus(void* responseHeaderMap, int status) {
  ASSERT(nullptr != responseHeaderMap);

  auto that = static_cast<Envoy::Http::ResponseHeaderMap*>(responseHeaderMap);
  that->setStatus(status);
}
