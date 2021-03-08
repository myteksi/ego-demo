#pragma once
// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#ifndef CGO_GOC_H
#define CGO_GOC_H

#include "envoy/http/filter.h"
#include "envoy/stats/stats.h"
#include "envoy/stream_info/filter_state.h"

// FilterStatus conversion
Envoy::Http::FilterHeadersStatus Goc_FilterHeadersStatus(int status);
Envoy::Http::FilterTrailersStatus Goc_FilterTrailersStatus(int status);
Envoy::Http::FilterDataStatus Goc_FilterDataStatus(int status);
// Http status code conversion
Envoy::Http::Code Goc_HttpResonseCode(int responseCode);
// FilterState enums conversion
Envoy::StreamInfo::FilterState::StateType Goc_FilterStateType(int stateType);
Envoy::StreamInfo::FilterState::LifeSpan Goc_FilterStateLifeSpan(int lifeSpan);
// Stats enums conversion
Envoy::Stats::Gauge::ImportMode Goc_Stats_ImportMode(int importMode);
Envoy::Stats::Histogram::Unit Goc_Stats_Unit(int unit);
int Goc_Stats_Unit_Value(Envoy::Stats::Histogram::Unit unit);

#endif // CGO_GOC_H
