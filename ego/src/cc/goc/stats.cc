// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include "envoy/stats/scope.h"

#include "common/stats/symbol_table_impl.h"

#include "envoy.h"
#include "goc.h"

// Stats::Scope
//
const void* Stats_Scope_counterFromStatName(void* scope, GoStr name) {
  auto ptr = static_cast<Envoy::Stats::Scope*>(scope);
  auto w_name = absl::string_view(name.data, name.len);

  Envoy::Stats::StatNameManagedStorage storage(w_name, ptr->symbolTable());
  Envoy::Stats::StatName stat_name = storage.statName();

  auto c = &ptr->counterFromStatName(stat_name);
  return c;
}

const void* Stats_Scope_gaugeFromStatName(void* scope, GoStr name, int importMode) {
  auto ptr = static_cast<Envoy::Stats::Scope*>(scope);
  auto w_name = absl::string_view(name.data, name.len);

  Envoy::Stats::StatNameManagedStorage storage(w_name, ptr->symbolTable());
  Envoy::Stats::StatName stat_name = storage.statName();

  auto c = &ptr->gaugeFromStatName(stat_name, Goc_Stats_ImportMode(importMode));
  return c;
}

const void* Stats_Scope_histogramFromStatName(void* scope, GoStr name, int unit) {
  auto ptr = static_cast<Envoy::Stats::Scope*>(scope);
  auto w_name = absl::string_view(name.data, name.len);

  Envoy::Stats::StatNameManagedStorage storage(w_name, ptr->symbolTable());
  Envoy::Stats::StatName stat_name = storage.statName();

  auto c = &ptr->histogramFromStatName(stat_name, Goc_Stats_Unit(unit));
  return c;
}

// Stats::Counter
//
void Stats_Counter_add(void* counter, uint64_t amount) {
  auto ptr = static_cast<Envoy::Stats::Counter*>(counter);
  ptr->add(amount);
}

void Stats_Counter_inc(void* counter) {
  auto ptr = static_cast<Envoy::Stats::Counter*>(counter);
  ptr->inc();
}

uint64_t Stats_Counter_latch(void* counter) {
  auto ptr = static_cast<Envoy::Stats::Counter*>(counter);
  return ptr->latch();
}
void Stats_Counter_reset(void* counter) {
  auto ptr = static_cast<Envoy::Stats::Counter*>(counter);
  ptr->reset();
}
uint64_t Stats_Counter_value(void* counter) {
  auto ptr = static_cast<Envoy::Stats::Counter*>(counter);
  return ptr->value();
}

// Stats::Gauge
//
void Stats_Gauge_add(void* gauge, uint64_t amount) {
  auto ptr = static_cast<Envoy::Stats::Gauge*>(gauge);
  ptr->add(amount);
}

void Stats_Gauge_dec(void* gauge) {
  auto ptr = static_cast<Envoy::Stats::Gauge*>(gauge);
  ptr->dec();
}

void Stats_Gauge_inc(void* gauge) {
  auto ptr = static_cast<Envoy::Stats::Gauge*>(gauge);
  ptr->inc();
}

void Stats_Gauge_set(void* gauge, uint64_t value) {
  auto ptr = static_cast<Envoy::Stats::Gauge*>(gauge);
  ptr->set(value);
}

void Stats_Gauge_sub(void* gauge, uint64_t amount) {
  auto ptr = static_cast<Envoy::Stats::Gauge*>(gauge);
  ptr->sub(amount);
}

uint64_t Stats_Gauge_value(void* gauge) {
  auto ptr = static_cast<Envoy::Stats::Gauge*>(gauge);
  return ptr->value();
}

// Stats::Histogram
int Stats_Histogram_unit(void* histogram) {
  auto ptr = static_cast<Envoy::Stats::Histogram*>(histogram);
  return Goc_Stats_Unit_Value(ptr->unit());
}

void Stats_Histogram_recordValue(void* histogram, uint64_t value) {
  auto ptr = static_cast<Envoy::Stats::Histogram*>(histogram);
  ptr->recordValue(value);
}