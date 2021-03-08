// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "test/test_common/utility.h"

#include "ego/src/cc/filter/http/filter.h"
#include "ego/src/cc/goc/envoy.h"
#include "ego/src/cc/goc/proto/dto.pb.validate.h"

class RequestHeaderMapGetByPrefixTest : public testing::Test {
public:
  void init(char* prefix, char* buffer, size_t buffer_size) {
    EXPECT_EQ(strlen(prefix), 5);
    prefix_.len = strlen(prefix);
    prefix_.data = prefix;

    buffer_.data = buffer;
    buffer_.len = buffer_size;
    buffer_.cap = buffer_size;
  }

  GoBuf buffer_;
  GoStr prefix_;
};

TEST_F(RequestHeaderMapGetByPrefixTest, FoundHeaders) {
  Envoy::Http::TestRequestHeaderMapImpl request_headers{{"x-ego1", "val1"}, {"x-ego2", "val2"}};

  char c_buffer[100];
  char c_prefix[] = "x-ego";
  init(c_prefix, c_buffer, 100);

  auto size = RequestHeaderMap_getByPrefix(&request_headers, prefix_, buffer_);
  EXPECT_LT(0, size);

  auto result = ego::http::RequestHeaderMap{};
  auto ok = result.ParseFromArray(buffer_.data, size);
  ASSERT_TRUE(ok);

  auto expected_result = ego::http::RequestHeaderMap{};

  auto x_ego1_header = expected_result.add_headers();
  x_ego1_header->set_key("x-ego1");
  x_ego1_header->set_value("val1");

  auto x_ego2_header = expected_result.add_headers();
  x_ego2_header->set_key("x-ego2");
  x_ego2_header->set_value("val2");

  ASSERT_EQ(expected_result.SerializePartialAsString(), result.SerializePartialAsString());
}

TEST_F(RequestHeaderMapGetByPrefixTest, NotFoundHeaders) {
  Envoy::Http::TestRequestHeaderMapImpl request_headers{};

  char c_buffer[100];
  char c_prefix[] = "x-ego";
  init(c_prefix, c_buffer, 100);

  auto size = RequestHeaderMap_getByPrefix(&request_headers, prefix_, buffer_);
  EXPECT_EQ(0, size);
}

TEST_F(RequestHeaderMapGetByPrefixTest, ShouldNotWriteOverBufferSize) {
  Envoy::Http::TestRequestHeaderMapImpl request_headers{{"x-ego1", "val1"}, {"x-ego2", "val2"}};

  const auto buffer_size = 34;
  char c_buffer[] = "0123456789012345678901234567890123should_not_be_overriden";

  char c_prefix[] = "x-ego";
  init(c_prefix, c_buffer, buffer_size);

  auto size = RequestHeaderMap_getByPrefix(&request_headers, prefix_, buffer_);
  EXPECT_EQ(32, size);

  // getByPrefix should not write over the buffer size
  char expected_buffer[] = "0123456789012345678901234567890123should_not_be_overriden";
  ASSERT_STREQ(expected_buffer + buffer_size, c_buffer + buffer_size);
}

TEST_F(RequestHeaderMapGetByPrefixTest, TooSmallBuffer) {
  Envoy::Http::TestRequestHeaderMapImpl request_headers{{"x-ego1", "val1"}, {"x-ego2", "val2"}};

  const auto buffer_size = 10;
  char c_buffer[] = "this is a buffer with existing data";

  char c_prefix[] = "x-ego";
  init(c_prefix, c_buffer, buffer_size);

  auto size = RequestHeaderMap_getByPrefix(&request_headers, prefix_, buffer_);
  EXPECT_GT(size, buffer_size);

  // getByPrefix should not write to buffer size
  char expected_buffer[] = "this is a buffer with existing data";
  ASSERT_STREQ(expected_buffer, c_buffer);
}