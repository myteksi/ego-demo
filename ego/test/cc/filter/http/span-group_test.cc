// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "test/mocks/http/mocks.h"
#include "test/mocks/tracing/mocks.h"
#include "test/test_common/utility.h"
#include "ego/src/cc/filter/http/span-group.h"
#include "common/tracing/http_tracer_impl.h"


using testing::_;
using testing::NiceMock;
using testing::Return;
using testing::ByMove;
using testing::AtLeast;

namespace Envoy {
namespace Http {

class SpanGroupTest : public testing::Test {
  public:
    void initialize(){
      span_group_.setDecoderFilterCallbacks(decoder_callbacks_);
      span_group_.setEncoderFilterCallbacks(encoder_callbacks_);
      EXPECT_CALL(decoder_callbacks_, activeSpan()).WillRepeatedly(ReturnRef(decoder_active_span_));
      EXPECT_CALL(encoder_callbacks_, activeSpan()).WillRepeatedly(ReturnRef(encoder_active_span_));
    };

    SpanGroup span_group_;
    NiceMock<Envoy::Http::MockStreamDecoderFilterCallbacks> decoder_callbacks_;
    NiceMock<Envoy::Http::MockStreamEncoderFilterCallbacks> encoder_callbacks_;
    NiceMock<Tracing::MockSpan> decoder_active_span_;
    NiceMock<Tracing::MockSpan> encoder_active_span_;
};

TEST_F(SpanGroupTest, ReturnsActiveSpanFromDecoderCallbacksIfSpanIDIsMinusOne) {
  initialize();

  EXPECT_EQ(&decoder_callbacks_.activeSpan(), &span_group_.getSpan(SpanGroupConstants::get().DecoderActiveSpan));
}

TEST_F(SpanGroupTest, ReturnsActiveSpanFromEncoderCallbacksIfSpanIDIsZero) {
  initialize();

  EXPECT_EQ(&encoder_callbacks_.activeSpan(), &span_group_.getSpan(SpanGroupConstants::get().EncoderActiveSpan));
}

TEST_F(SpanGroupTest, ReturnsNullSpanIfNotFoundSpanID) {
  initialize();

  EXPECT_EQ(&Envoy::Tracing::NullSpan::instance(), &span_group_.getSpan(10));
}

TEST_F(SpanGroupTest, ReturnsNullSpanIfDecoderCallbacksIsNull) {
  EXPECT_EQ(&Envoy::Tracing::NullSpan::instance(), &span_group_.getSpan(SpanGroupConstants::get().DecoderActiveSpan));
}

TEST_F(SpanGroupTest, ReturnsNullSpanIfEncoderCallbacksIsNull) {
  EXPECT_EQ(&Envoy::Tracing::NullSpan::instance(), &span_group_.getSpan(SpanGroupConstants::get().EncoderActiveSpan));
}

TEST_F(SpanGroupTest, SpawnChildFromDecoderActiveSpan) {
  initialize();

  Tracing::MockSpan* child_span{new Tracing::MockSpan()};
  EXPECT_CALL(decoder_active_span_, spawnChild_).WillOnce(Return(child_span));

  auto name = std::string("name");
  auto span_id = span_group_.spawnChildSpan(SpanGroupConstants::get().DecoderActiveSpan, name);

  EXPECT_CALL(*child_span, finishSpan).Times(AtLeast(1));
  span_group_.getSpan(span_id).finishSpan();
}

TEST_F(SpanGroupTest, SpawnChildFromEncoderActiveSpan) {
  initialize();

  Tracing::MockSpan* child_span{new Tracing::MockSpan()};
  EXPECT_CALL(encoder_active_span_, spawnChild_).WillOnce(Return(child_span));

  auto name = std::string("name");
  auto span_id = span_group_.spawnChildSpan(SpanGroupConstants::get().EncoderActiveSpan, name);

  EXPECT_CALL(*child_span, finishSpan).Times(AtLeast(1));
  span_group_.getSpan(span_id).finishSpan();
}

TEST_F(SpanGroupTest, SpawnChildFromExistingSpan) {
  initialize();

  Tracing::MockSpan* child_span{new Tracing::MockSpan()};
  EXPECT_CALL(decoder_active_span_, spawnChild_).WillOnce(Return(child_span));

  auto child_name = std::string("child_name");
  auto child_span_id = span_group_.spawnChildSpan(SpanGroupConstants::get().DecoderActiveSpan, child_name);

  Tracing::MockSpan* grand_child_span{new Tracing::MockSpan()};
  EXPECT_CALL(*child_span, spawnChild_).WillOnce(Return(grand_child_span)); 

  auto grand_child_name = std::string("grand_child_name");
  auto grand_child_span_id = span_group_.spawnChildSpan(child_span_id, grand_child_name);

  EXPECT_CALL(*grand_child_span, finishSpan).Times(AtLeast(1));
  span_group_.getSpan(grand_child_span_id).finishSpan();
}

TEST_F(SpanGroupTest, SpawnChildFromNullSpan) {
  initialize();

  auto name = std::string("something");
  auto child_span_id = span_group_.spawnChildSpan(100, name);
  
  span_group_.getSpan(child_span_id).finishSpan();
}

TEST_F(SpanGroupTest, DeleteSpan) {
  initialize();

  Tracing::MockSpan* child_span{new Tracing::MockSpan()};
  EXPECT_CALL(encoder_active_span_, spawnChild_).WillOnce(Return(child_span));

  auto name = std::string("something");
  auto child_span_id = span_group_.spawnChildSpan(SpanGroupConstants::get().EncoderActiveSpan, name);
  span_group_.removeSpan(child_span_id);
  EXPECT_EQ(&Envoy::Tracing::NullSpan::instance(), &span_group_.getSpan(child_span_id));
}

TEST_F(SpanGroupTest, DeleteNotExistingSpan) {
  initialize();
  span_group_.removeSpan(123);
}

}
}