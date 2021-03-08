// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "test/mocks/http/mocks.h"
#include "test/test_common/utility.h"

#include "ego/src/cc/filter/http/filter.h"
#include "mocks.h"

using testing::_;
using testing::NiceMock;
using testing::Return;

namespace Envoy {
namespace Http {

class GoHttpRouteSpecificFilterConfigTest : public testing::Test {
public:
  void initializeConfig(const std::string& yaml) {
    auto config = TestUtility::parseYaml<ego::http::SettingsPerRoute>(yaml);
    config_ = std::make_shared<GoHttpRouteSpecificFilterConfig>(config);
  }

  std::shared_ptr<GoHttpRouteSpecificFilterConfig> config_;
};

TEST_F(GoHttpRouteSpecificFilterConfigTest, MultipleFilterConfigurations) {
  // NOTE: the getheader example is not accurate -- the filter does not support
  //       route specific configuration. But it should pass.
  const std::string config_yaml = R"EOF(
    filters:
      getheader:
        "@type": type.googleapis.com/egodemo.getheader.Settings
        key: "x-123"
        src: "http://example.com"
        hdr: "x-yz"
      security:
        "@type": type.googleapis.com/ego.security.Requirement
        provider_name: hmac
    )EOF";

  initializeConfig(config_yaml);

  ASSERT_NE(config_, nullptr);

  auto cgoTagGetHeader = config_->cgoTag("getheader");
  auto cgoTagSecurity = config_->cgoTag("security");
  EXPECT_NE(0, cgoTagGetHeader);
  EXPECT_NE(0, cgoTagSecurity);
  EXPECT_NE(cgoTagGetHeader, cgoTagSecurity);
}

TEST_F(GoHttpRouteSpecificFilterConfigTest, EmptyConfiguration) {
  const std::string config_yaml = R"EOF(
    filters:
    )EOF";

  initializeConfig(config_yaml);

  ASSERT_NE(config_, nullptr);
  EXPECT_EQ(0, config_->cgoTag("getheader"));
}

class GoHttpFilterResolveMostSpecificPerGoFilterConfigTagTest : public testing::Test {
public:
  void initializeFilter(std::string config_yaml) {
    auto settings = TestUtility::parseYaml<ego::http::Settings>(config_yaml);
    stats_scope_ = std::make_shared<Stats::IsolatedStoreImpl>();
    auto config = std::make_shared<GoHttpFilterConfig>(settings, stats_scope_);

    auto api = Envoy::Api::createApiForTest();
    auto cgo_proxy = std::make_shared<CgoProxyImpl>();

    filter_ = new GoHttpFilter(config, *api, nullptr, cgo_proxy, std::make_unique<SpanGroup>());
    stream_filter_ = filter_->ref();

    filter_->setDecoderFilterCallbacks(decoder_callbacks_);
    filter_->setEncoderFilterCallbacks(encoder_callbacks_);
  }

  std::shared_ptr<GoHttpRouteSpecificFilterConfig>
  createRouteSpecificFilterConfig(std::string config_yaml) {
    auto route_config = TestUtility::parseYaml<ego::http::SettingsPerRoute>(config_yaml);
    return std::make_shared<GoHttpRouteSpecificFilterConfig>(route_config);
  }

  void cleanUp() { filter_->onDestroy(); }

  GoHttpFilter* filter_;
  Envoy::Http::StreamFilterSharedPtr stream_filter_;
  NiceMock<Envoy::Http::MockStreamDecoderFilterCallbacks> decoder_callbacks_;
  NiceMock<Envoy::Http::MockStreamEncoderFilterCallbacks> encoder_callbacks_;
  Envoy::Stats::ScopeSharedPtr stats_scope_;
};

TEST_F(GoHttpFilterResolveMostSpecificPerGoFilterConfigTagTest, ValidConfig) {
  const std::string config_yaml = R"EOF(
    filter: security
    )EOF";
  initializeFilter(config_yaml);

  const std::string route_specific_config_yaml = R"EOF(
    filters:
      security:
        "@type": type.googleapis.com/ego.security.Requirement
        provider_name: hmac
    )EOF";
  auto route_config = createRouteSpecificFilterConfig(route_specific_config_yaml);
  EXPECT_CALL(*decoder_callbacks_.route_, perFilterConfig(GoHttpConstants::get().FilterName))
      .WillOnce(Return(route_config.get()));

  auto cgo_tag = filter_->resolveMostSpecificPerGoFilterConfigTag();
  EXPECT_NE(0, cgo_tag);

  cleanUp();
}

TEST_F(GoHttpFilterResolveMostSpecificPerGoFilterConfigTagTest, NotExistingConfig) {
  const std::string config_yaml = R"EOF(
    filter: security
    )EOF";
  initializeFilter(config_yaml);

  const std::string route_specific_config_yaml = R"EOF(
    filters:
      getheader:
        "@type": type.googleapis.com/egodemo.getheader.Settings
        key: "x-123"
        src: "http://example.com"
        hdr: "x-yz"
    )EOF";
  auto route_config = createRouteSpecificFilterConfig(route_specific_config_yaml);

  EXPECT_CALL(*decoder_callbacks_.route_, perFilterConfig(GoHttpConstants::get().FilterName))
      .WillOnce(Return(route_config.get()));

  auto cgo_tag = filter_->resolveMostSpecificPerGoFilterConfigTag();
  EXPECT_EQ(0, cgo_tag);

  cleanUp();
}

TEST_F(GoHttpFilterResolveMostSpecificPerGoFilterConfigTagTest, NullRouteSpecificConfig) {
  const std::string config_yaml = R"EOF(
    filter: security
    )EOF";
  initializeFilter(config_yaml);

  EXPECT_CALL(*decoder_callbacks_.route_, perFilterConfig(GoHttpConstants::get().FilterName))
      .WillOnce(Return(nullptr));

  auto cgo_tag = filter_->resolveMostSpecificPerGoFilterConfigTag();
  EXPECT_EQ(0, cgo_tag);

  cleanUp();
}

TEST_F(GoHttpFilterResolveMostSpecificPerGoFilterConfigTagTest, NullRoute) {
  const std::string config_yaml = R"EOF(
    filter: security
    )EOF";
  initializeFilter(config_yaml);

  EXPECT_CALL(decoder_callbacks_, route).WillOnce(Return(nullptr));

  auto cgo_tag = filter_->resolveMostSpecificPerGoFilterConfigTag();
  EXPECT_EQ(0, cgo_tag);

  cleanUp();
}

TEST_F(GoHttpFilterResolveMostSpecificPerGoFilterConfigTagTest, NullRouteEntry) {
  const std::string config_yaml = R"EOF(
    filter: security
    )EOF";
  initializeFilter(config_yaml);

  EXPECT_CALL(*decoder_callbacks_.route_, routeEntry).WillOnce(Return(nullptr));

  auto cgo_tag = filter_->resolveMostSpecificPerGoFilterConfigTag();
  EXPECT_EQ(0, cgo_tag);

  cleanUp();
}

class GoHttpFilterTest : public testing::Test {
public:
  void initializeFilter() {
    const std::string config_yaml = R"EOF(
        filter: security
      )EOF";
    auto settings = TestUtility::parseYaml<ego::http::Settings>(config_yaml);
    stats_scope_ = std::make_shared<Stats::IsolatedStoreImpl>();
    auto config = std::make_shared<GoHttpFilterConfig>(settings, stats_scope_);
    auto api = Envoy::Api::createApiForTest();

    cgo_proxy_ = std::make_shared<NiceMock<MockCgoProxy>>();
    EXPECT_CALL(*cgo_proxy_, GoHttpFilterCreate).WillOnce(Return(100));

    filter_ = new GoHttpFilter(config, *api, nullptr, cgo_proxy_, std::make_unique<SpanGroup>());
    stream_filter_ = filter_->ref();

    filter_->setDecoderFilterCallbacks(decoder_callbacks_);
    filter_->setEncoderFilterCallbacks(encoder_callbacks_);
  }

  void cleanUp() { filter_->onDestroy(); }

  GoHttpFilter* filter_;
  Envoy::Http::StreamFilterSharedPtr stream_filter_;
  NiceMock<Envoy::Http::MockStreamDecoderFilterCallbacks> decoder_callbacks_;
  NiceMock<Envoy::Http::MockStreamEncoderFilterCallbacks> encoder_callbacks_;
  std::shared_ptr<MockCgoProxy> cgo_proxy_;
  Envoy::Stats::ScopeSharedPtr stats_scope_;
};

TEST_F(GoHttpFilterTest, DecodeHeaders) {
  initializeFilter();

  Http::TestRequestHeaderMapImpl request_headers;
  EXPECT_CALL(*cgo_proxy_, GoHttpFilterDecodeHeaders(_, &request_headers, true));

  filter_->decodeHeaders(request_headers, true);

  cleanUp();
}

TEST_F(GoHttpFilterTest, DecodeData) {
  initializeFilter();

  Buffer::OwnedImpl data;
  EXPECT_CALL(*cgo_proxy_, GoHttpFilterDecodeData(_, &data, true));

  filter_->decodeData(data, true);

  cleanUp();
}

TEST_F(GoHttpFilterTest, DecodeTrailers) {
  initializeFilter();

  Http::TestRequestTrailerMapImpl request_trailers;
  EXPECT_CALL(*cgo_proxy_, GoHttpFilterDecodeTrailers(_, &request_trailers));

  filter_->decodeTrailers(request_trailers);

  cleanUp();
}

TEST_F(GoHttpFilterTest, EncodeHeaders) {
  initializeFilter();

  Http::TestResponseHeaderMapImpl response_headers;
  EXPECT_CALL(*cgo_proxy_, GoHttpFilterEncodeHeaders(_, &response_headers, true));

  filter_->encodeHeaders(response_headers, true);

  cleanUp();
}

TEST_F(GoHttpFilterTest, EncodeData) {
  initializeFilter();

  Buffer::OwnedImpl data;
  EXPECT_CALL(*cgo_proxy_, GoHttpFilterEncodeData(_, &data, false));

  filter_->encodeData(data, false);

  cleanUp();
}

TEST_F(GoHttpFilterTest, StreamFilterCallbacksWithFalseEncoder) {
  initializeFilter();

  auto *callbacks = filter_->streamFilterCallbacks(0);

  EXPECT_EQ(&decoder_callbacks_, callbacks);

  cleanUp();
}

TEST_F(GoHttpFilterTest, StreamFilterCallbacksWithTrueEncoder) {
  initializeFilter();

  auto *callbacks = filter_->streamFilterCallbacks(1);

  EXPECT_EQ(&encoder_callbacks_, callbacks);

  cleanUp();
}

TEST_F(GoHttpFilterTest, PostDownCallFlow) {
  initializeFilter();

  auto post_tag = 1;
  EXPECT_CALL(decoder_callbacks_.dispatcher_, post(_)).WillOnce([](std::function<void()> callback) {
    callback();
  });
  EXPECT_CALL(*cgo_proxy_, GoHttpFilterOnPost);

  filter_->pin();
  filter_->post(post_tag);
  filter_->unpin();

  cleanUp();
}

} // namespace Http
} // namespace Envoy