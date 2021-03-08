// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "envoy/http/header_map.h"

#include "absl/strings/match.h"
#include "ego/src/cc/goc/proto/dto.pb.validate.h"
#include "envoy.h"

void RequestHeaderMap_add(void* requestHeaderMap, GoStr name, GoStr value) {
  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

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

void RequestHeaderMap_set(void* requestHeaderMap, GoStr name, GoStr value) {
  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

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

void RequestHeaderMap_append(void* requestHeaderMap, GoStr name, GoStr value) {
  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

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

size_t RequestHeaderMap_getByPrefix(void* requestHeaderMap, GoStr prefix, GoBuf buf) {
  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

  auto c_prefix = std::string(prefix.data, prefix.len);

  auto result = ego::http::RequestHeaderMap{};
  auto args = std::make_pair(c_prefix, &result);

  that->iterate(
      [](const Envoy::Http::HeaderEntry& header, void* context) -> Envoy::Http::HeaderMap::Iterate {
        auto key_ret = static_cast<std::pair<std::string, ego::http::RequestHeaderMap*>*>(context);
        const absl::string_view header_key_view = header.key().getStringView();
        if (absl::StartsWith(header_key_view, key_ret->first)) {
          auto entry = std::make_unique<ego::http::HeaderEntry>();
          entry->set_key(header.key().getStringView().data(), header.key().getStringView().size());
          entry->set_value(header.value().getStringView().data(),
                           header.value().getStringView().size());
          key_ret->second->mutable_headers()->Add(std::move(*entry));
        }
        return Envoy::Http::HeaderMap::Iterate::Continue;
      },
      &args);

  // no headers matches the prefix
  if (0 == result.headers_size()) {
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

void RequestHeaderMap_remove(void* requestHeaderMap, GoStr name) {
  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

  // The std::string() constructor will create a copy of name. Unfortunately,
  // there is no std::string_view, because...
  auto c_name = std::string(name.data, name.len);

  // ...LowerCaseString creates a wrapped lowercase copy of c_name.
  auto w_name = Envoy::Http::LowerCaseString(c_name);
  that->remove(w_name);
}

void RequestHeaderMap_Path(void* requestHeaderMap, GoStr* value) {
  ASSERT(nullptr != requestHeaderMap);
  ASSERT(nullptr != value);

  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

  ASSERT(nullptr != that->Path());
  auto path = that->Path()->value().getStringView();

  // Path() returns a pointer, value() returns a reference,
  // therefore getStringView().data() should be valid after return
  value->len = path.size();
  value->data = const_cast<char*>(path.data());
}

// changing path might update result of resolveMostSpecificPerFilterConfig
// if route cache is cleared
void RequestHeaderMap_setPath(void* requestHeaderMap, GoStr path) {
  ASSERT(nullptr != requestHeaderMap);

  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

  // we are wrapping the header value with a lightweight string_view, which
  // is safe because...
  auto w_value = absl::string_view(path.data, path.len);

  that->setPath(w_value);
}

void RequestHeaderMap_Method(void* requestHeaderMap, GoStr* value) {
  ASSERT(nullptr != requestHeaderMap);
  ASSERT(nullptr != value);

  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

  ASSERT(nullptr != that->Method());
  auto method = that->Method()->value().getStringView();

  // Method() returns a pointer, value() returns a reference,
  // therefore getStringView().data() should be valid after return
  value->len = method.size();
  value->data = const_cast<char*>(method.data());
}

void RequestHeaderMap_ContentType(void* requestHeaderMap, GoStr* value) {
  ASSERT(nullptr != requestHeaderMap);
  ASSERT(nullptr != value);

  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

  if (nullptr == that->ContentType()) {
    return;
  }
  auto contentType = that->ContentType()->value().getStringView();

  // ContentType() returns a pointer, value() returns a reference,
  // therefore getStringView().data() should be valid after return
  value->len = contentType.size();
  value->data = const_cast<char*>(contentType.data());
}

void RequestHeaderMap_Authorization(void* requestHeaderMap, GoStr* value) {
  ASSERT(nullptr != requestHeaderMap);
  ASSERT(nullptr != value);

  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

  if (that->Authorization() == nullptr) {
    return;
  }
  auto authorization = that->Authorization()->value().getStringView();

  // Authorization() returns a pointer, value() returns a reference,
  // therefore getStringView().data() should be valid after return
  value->len = authorization.size();
  value->data = const_cast<char*>(authorization.data());
}

void RequestHeaderMap_get(void* requestHeaderMap, GoStr key, GoStr* value) {
  ASSERT(nullptr != requestHeaderMap);
  ASSERT(nullptr != value);

  auto that = static_cast<Envoy::Http::RequestHeaderMap*>(requestHeaderMap);

  auto cName = std::string(key.data, key.len);
  auto wName = Envoy::Http::LowerCaseString(cName);

  if (that->get(wName) == nullptr) {
    return;
  }

  auto valStringView = that->get(wName)->value().getStringView();

  // get() returns a pointer, value() returns a reference,
  // therefore getStringView().data() should be valid after return
  value->len = valStringView.size();
  value->data = const_cast<char*>(valStringView.data());
}
